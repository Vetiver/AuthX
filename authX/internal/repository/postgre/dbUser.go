package postgre

import (
	"authX/internal/domain"
	"authX/internal/kafka"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

func (db *DB) CreateUserWithEvent(ctx context.Context, user *domain.User, event kafka.AuditEvent) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO users (email, role, password, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id`
	err = tx.QueryRow(ctx, query, user.Email, user.Role, user.Password).Scan(&user.ID)
	if err != nil {
		db.logger.Error("Failed to create user", zap.Error(err))
		return fmt.Errorf("create user: %w", err)
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO outbox_events (event_type, payload) VALUES ($1, $2)", event.EventType, payload)
	if err != nil {
		return fmt.Errorf("save outbox event: %w", err)
	}

	return tx.Commit(ctx)
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

func (db *DB) SaveAuditEvent(ctx context.Context, event kafka.AuditEvent) error {
	query := `
		INSERT INTO audit_events (id, event_type, occurred_at, user_id, email, role, ip, user_agent, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
	`

	_, err := db.pool.Exec(ctx, query,
		event.EventID,
		event.EventType,
		event.OccurredAt,
		event.UserID,
		event.Email,
		event.Role,
		event.IP,
		event.UserAgent,
		event.Metadata,
	)
	if err != nil {
		db.logger.Error("Failed to save audit event", zap.Error(err))
		return fmt.Errorf("save audit event: %w", err)
	}

	return nil
}
