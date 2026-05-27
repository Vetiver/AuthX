package handlers

import (
	"authX/authX/internal/domain"
	"authX/authX/utils/config"
	"authX/authX/utils/constants"
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DomainService interface {
	RegisterUser(ctx context.Context, dto domain.RegisterUserDto) error
	Login(ctx context.Context, dto domain.LoginUserDto) (*domain.UserLoginResp, error)
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
