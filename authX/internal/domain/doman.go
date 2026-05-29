package domain

import (
	"context"
	"time"

)

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Role      string    `json:"role,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterUserDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserDto struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResp struct {
	User
	Token string `json:"token"`
}

type ValidateResponse struct {
	Valid  bool   `json:"valid"`
	UserID int `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type PostgreRepo interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type RedisRepo interface {
	SaveUserToken(ctx context.Context, userID int, token string, ttl time.Duration) error
	GetUserToken(ctx context.Context, userID int) (string, error)
	AddToBlacklist(ctx context.Context, tokenID string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
}