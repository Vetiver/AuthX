package transport

import (
	"context"
	"net/http"

	"authX/internal/middleware"
	"authX/pkg"
	"authX/transport/handlers"
	"authX/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RedisRepo interface {
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
}

type HttpServer struct {
	logger     *zap.Logger
	httpPort   string
	redisDB    RedisRepo
	jwtManager *utils.JWTManager
}

func NewHttpServer(logger *zap.Logger, httpPort string, redisDB RedisRepo, jwtManager *utils.JWTManager) *HttpServer {
	return &HttpServer{
		logger:     logger,
		httpPort:   httpPort,
		redisDB:    redisDB,
		jwtManager: jwtManager,
	}
}

func (h *HttpServer) StartHTTPServer(handlers *handlers.BaseHandler) *http.Server {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(pkg.AccessLog())

	router.GET("/ping", handlers.Ping)

	auth := router.Group("/auth")
	{
		auth.POST("/login", handlers.Login)
		auth.POST("/register", handlers.Register)
	}

	authProtected := router.Group("/auth")
	authProtected.Use(middleware.AuthMiddleware(h.jwtManager, h.redisDB, h.logger))
	{
		authProtected.GET("/roles", handlers.GetRoles)
	}

	srv := &http.Server{
		Addr:    h.httpPort,
		Handler: router,
	}

	go func() {
		h.logger.Info("HTTP server is running on port", zap.String("port", h.httpPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Fatal("failed to start HTTP server", zap.Error(err))
		}
	}()

	return srv
}