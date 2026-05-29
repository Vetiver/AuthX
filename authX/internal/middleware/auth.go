package middleware

import (
	"context"
	"net/http"
	"strings"

	"authX/authX/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RedisRepo interface {
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
}

func AuthMiddleware(jwtManager *utils.JWTManager, redisDB RedisRepo, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "NO_TOKEN",
				"message": "authorization header required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_FORMAT",
				"message": "invalid authorization format",
			})
			return
		}

		claims, err := jwtManager.Validate(parts[1])
		if err != nil {
			logger.Warn("Invalid token", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "TOKEN_INVALID",
				"message": "token invalid or expired",
			})
			return
		}

		isBlacklisted, err := redisDB.IsBlacklisted(c.Request.Context(), claims.TokenID)
		if err != nil {
			logger.Error("Blacklist check failed", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "internal error",
			})
			return
		}
		if isBlacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "TOKEN_REVOKED",
				"message": "token has been revoked",
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}