package maim

import (
	"authX/authX/internal/domain"
	"authX/authX/internal/repository/postgre"
	"authX/authX/internal/repository/redis"
	"authX/authX/pkg"
	"authX/authX/transport"
	"authX/authX/transport/handlers"
	"authX/authX/utils"
	"authX/authX/utils/config"
	"context"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	logger := pkg.CreateLogger()
	defer logger.Sync()
	ctx := context.Background() 
	wg := sync.WaitGroup{}

	wg.Add(1)
	config := config.NewConfig()
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
	httpServer := transport.NewHttpServer(logger, config.HTTPAddr)
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