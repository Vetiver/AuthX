package transport

import (
	"authX/authX/pkg"
	"authX/authX/transport/handlers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HttpServer struct {
	logger   *zap.Logger
	httpPort string
}

func NewHttpServer(logger *zap.Logger, httpPort string) *HttpServer {
	return &HttpServer{
		logger:   logger,
		httpPort: httpPort,
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
		// auth.POST("/logout", handlers.Logout)
		// auth.POST("/refresh", handlers.RefreshToken)
	}

	h.logger.Info("HTTP server is running on port", zap.String("port", h.httpPort))

	if err := router.Run(":" + h.httpPort); err != nil {
		h.logger.Fatal("failed to start HTTP server", zap.Error(err))
	}
}
