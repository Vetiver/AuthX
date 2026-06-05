package main

import (
	"authX/internal/domain"
	"authX/internal/kafka"
	"authX/internal/repository/postgre"
	"authX/internal/repository/redis"
	"authX/internal/scheduler"
	"authX/pkg"
	"authX/transport"
	"authX/transport/handlers"
	"authX/utils"
	"authX/utils/config"
	"context"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	logger := pkg.CreateLogger()
	defer logger.Sync()
	ctx := context.Background() 
	wg := sync.WaitGroup{}
	config := config.NewConfig()
	brokers := strings.Split(config.KafkaBrokers, ",")
	producer := kafka.NewProducer(brokers, config.KafkaAuthEventsTopic, logger)
	defer producer.Close()

	wg.Add(1)
	pool := postgre.DbStart(config, logger)
	if pool == nil {
		logger.Fatal("Failed to connect to the database")
		return
	}
	database := postgre.NewRepository(pool, logger)
	if database == nil {
		logger.Fatal("Failed to start database")
		return
	}
	redisDB, err := redis.NewRedisDB(ctx, config.RedisAddr, config.RedisPassword, config.RedisDB, logger)
	if err != nil {
		logger.Fatal("Failed to start redis")
		return
	}
	jwtManager := utils.NewJWTManager(config.JWTSecret, config.JWTAccessTTL)

	hasher := utils.NewBcryptHasher(bcrypt.DefaultCost)
	domainService := domain.NewDomainService(logger, config, database, hasher, jwtManager, redisDB)
	httpHandlers := handlers.NewBaseHandler(logger, domainService, config)
	httpServer := transport.NewHttpServer(logger, config.HTTPAddr, redisDB, jwtManager)
	outboxScheduler := scheduler.NewOutboxScheduler(database, producer, logger)
	go outboxScheduler.Start(ctx)
	// utils.StartPprofServer(":6066")
	// done := make(chan struct{})

	go func() {
		httpServer.StartHTTPServer(httpHandlers)
		logger.Error("HTTP server down")
		defer wg.Done()

		// done <- struct{}{}
	}()

	// <-done

	wg.Wait()
}