package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	kafka_repo "authX/internal/kafka"

	"authX/internal/repository/postgre"
	"authX/pkg"
	"authX/utils/config"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func main() {
	logger := pkg.CreateLogger()
	defer logger.Sync()

	cfg := config.NewConfig()

	pool := postgre.DbStart(cfg, logger)
	if pool == nil {
		logger.Fatal("Failed to connect to database")
	}
	db := postgre.NewRepository(pool, logger)

	brokers := strings.Split(cfg.KafkaBrokers, ",")
	logger.Info("Waiting for Kafka...")
	time.Sleep(10 * time.Second)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       cfg.KafkaAuthEventsTopic,
		GroupID:     cfg.KafkaAuditConsumerGroup,
		StartOffset: kafka.LastOffset,
		MinBytes:    10,
		MaxBytes:    10e6,
	})
	defer reader.Close()

	dlqWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        cfg.KafkaAuthEventsDLQTopic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
	}
	defer dlqWriter.Close()

	logger.Info("Audit consumer started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ctx := context.Background()

	go func() {
		for {
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				logger.Error("Read error", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			var event kafka_repo.AuditEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				logger.Error("Unmarshal error", zap.Error(err))
				dlqWriter.WriteMessages(ctx, kafka.Message{
					Value: msg.Value,
					Headers: []kafka.Header{
						{Key: "error", Value: []byte(err.Error())},
					},
				})
				continue
			}

			var saveErr error
			for retry := 0; retry < 3; retry++ {
				saveErr = db.SaveAuditEvent(ctx, event)
				if saveErr == nil {
					break
				}
				logger.Error("Save error, retrying...",
					zap.Int("attempt", retry+1),
					zap.Error(saveErr),
				)
				time.Sleep(time.Second)
			}

			if saveErr != nil {
				logger.Error("Save failed after retries", zap.Error(saveErr))
				dlqWriter.WriteMessages(ctx, kafka.Message{
					Value: msg.Value,
				})
				continue
			}

			logger.Info("Saved", zap.String("event", event.EventType))
		}
	}()

	<-quit
	logger.Info("Shutting down...")
	reader.Close()
	dlqWriter.Close()
}