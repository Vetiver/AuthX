// internal/repository/postgre/outbox.go
package postgre

import (
	"context"
	"encoding/json"
	"fmt"

	"authX/internal/kafka"

	"go.uber.org/zap"
)

func (db *DB) SaveOutboxEvent(ctx context.Context, event kafka.AuditEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	_, err = db.pool.Exec(ctx,
		"INSERT INTO outbox_events (event_type, payload) VALUES ($1, $2)",
		event.EventType,
		payload,
	)
	return err
}

func (db *DB) GetAndPublishPendingEvents(ctx context.Context, limit int, producer *kafka.Producer, logger *zap.Logger) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx,
		"SELECT id, payload FROM outbox_events WHERE processed = false ORDER BY id LIMIT $1 FOR UPDATE SKIP LOCKED",
		limit,
	)
	if err != nil {
		return fmt.Errorf("query outbox: %w", err)
	}

	type outboxRow struct {
		id      int64
		payload []byte
	}
	var items []outboxRow

	for rows.Next() {
		var id int64
		var payloadBytes []byte

		if err := rows.Scan(&id, &payloadBytes); err != nil {
			rows.Close()
			return fmt.Errorf("scan outbox: %w", err)
		}
		items = append(items, outboxRow{id: id, payload: payloadBytes})
	}
	rows.Close()

	for _, item := range items {
		var event kafka.AuditEvent
		if err := json.Unmarshal(item.payload, &event); err != nil {
			logger.Error("Failed to unmarshal event", zap.Error(err))
			continue
		}

		if err := producer.Publish(ctx, event); err != nil {
			logger.Error("Failed to publish to Kafka", zap.Error(err))
			continue
		}

		if _, err := tx.Exec(ctx, "UPDATE outbox_events SET processed = true WHERE id = $1", item.id); err != nil {
			return fmt.Errorf("mark processed: %w", err)
		}
	}

	return tx.Commit(ctx)
}
