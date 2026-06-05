package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type AuditEvent struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	OccurredAt string          `json:"occurred_at"`
	UserID     int             `json:"user_id,omitempty"`
	Email      string          `json:"email"`
	Role       string          `json:"role,omitempty"`
	IP         string          `json:"ip,omitempty"`
	UserAgent  string          `json:"user_agent,omitempty"`
	Metadata   json.RawMessage `json:"metadata"`
}

type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

func NewProducer(brokers []string, topic string, logger *zap.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
	}

	return &Producer{
		writer: writer,
		logger: logger,
	}
}

func (p *Producer) Publish(ctx context.Context, event AuditEvent) error {
	event.EventID = uuid.New().String()
	event.OccurredAt = time.Now().UTC().Format(time.RFC3339)

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	key := []byte(fmt.Sprintf("%d", event.UserID))
	if event.UserID == 0 {
		key = []byte(event.Email)
	}

	msg := kafka.Message{
		Key:   key,
		Value: value,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "event_id", Value: []byte(event.EventID)},
			{Key: "source", Value: []byte("authx")},
			{Key: "content_type", Value: []byte("application/json")},
		},
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		p.logger.Error("Failed to publish event",
			zap.String("event_type", event.EventType),
			zap.Error(err),
		)
		return fmt.Errorf("publish event: %w", err)
	}

	p.logger.Info("Event published",
		zap.String("event_type", event.EventType),
		zap.String("email", event.Email),
	)

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
