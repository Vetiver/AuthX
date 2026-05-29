package postgre

import (
	"authX/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

func (r *DB) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, role, password, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Role,
		user.Password,
	)

	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err))
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *DB) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, role, password, created_at
		FROM users
		WHERE email = $1
	`

	return r.getUser(ctx, query, email)
}

func (r *DB) getUser(ctx context.Context, query string, args ...interface{}) (*domain.User, error) {
	var user domain.User

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.Role,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error("Failed to get user", zap.Error(err))
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &user, nil
}
