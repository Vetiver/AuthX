package handlers

import (
	"authX/internal/domain"
	"authX/utils/config"
	"authX/utils/constants"
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DomainService interface {
	RegisterUser(ctx context.Context, dto domain.RegisterUserDto) error
	Login(ctx context.Context, dto domain.LoginUserDto) (*domain.UserLoginResp, error)
	ValidateToken(ctx context.Context, tokenString string) (*domain.ValidateResponse, error)
	GetRoles() []string
}

type RegisterReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type BaseHandler struct {
	config        *config.Config
	logger        *zap.Logger
	domainService DomainService
	inProgress    sync.Map
}

func NewBaseHandler(logger *zap.Logger, domainService DomainService, config *config.Config) *BaseHandler {
	return &BaseHandler{
		config:        config,
		logger:        logger,
		domainService: domainService,
	}
}

func (h *BaseHandler) Ping(c *gin.Context) {
	h.logger.Info("Ping request received")
	c.String(http.StatusOK, constants.ServiceOK)
}

func (h *BaseHandler) Register(c *gin.Context) {
	var req RegisterReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	err := h.domainService.RegisterUser(c, domain.RegisterUserDto{Email: req.Email, Password: req.Password})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    "SUCCESS",
		"message": "user registered successfully",
	})
}

func (h *BaseHandler) Login(c *gin.Context) {
	var req LoginReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	resp, err := h.domainService.Login(c.Request.Context(), domain.LoginUserDto{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "LOGIN_FAILED",
			"message": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "SUCCESS",
		"data": resp,
	})
}

func (h *BaseHandler) Validate(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "NO_TOKEN",
			"message": "authorization header required",
		})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "INVALID_FORMAT",
			"message": "invalid authorization format",
		})
		return
	}

	resp, err := h.domainService.ValidateToken(c.Request.Context(), parts[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "TOKEN_INVALID",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "SUCCESS",
		"data": resp,
	})
}

func (h *BaseHandler) GetRoles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": "SUCCESS",
		"data": h.domainService.GetRoles(),
	})
}
