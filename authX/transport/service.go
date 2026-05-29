package transport

import (
	"authX/internal/middleware"
	"authX/pkg"
	"authX/transport/handlers"
	"authX/utils"
	"context"

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

func (h *HttpServer) StartHTTPServer(handlers *handlers.BaseHandler) {
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
		authProtected.GET("/validate")
		authProtected.GET("/roles", handlers.GetRoles)
	}

	h.logger.Info("HTTP server is running on port", zap.String("port", h.httpPort))

	if err := router.Run(":" + h.httpPort); err != nil {
		h.logger.Fatal("failed to start HTTP server", zap.Error(err))
	}
}
