package postgre

import (
	"context"
	"fmt"
)

func (db *DB) Migrate(ctx context.Context) error {
	//Для ролей наверное можно потом добавить ENUM
	migrations := map[string]string{
		"CREATE_AUDIT_EVENTS": `CREATE TABLE IF NOT EXISTS audit_events (
    	id UUID PRIMARY KEY,
    	event_type TEXT NOT NULL,
    	occurred_at TIMESTAMPTZ NOT NULL,
    	user_id INT NULL,
    	email TEXT NOT NULL,
    	role TEXT NULL,
    	ip TEXT NULL,
    	user_agent TEXT NULL,
    	metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	`,
		"CREATE_USERS": `CREATE TABLE IF NOT EXISTS users (
    	id SERIAL PRIMARY KEY,
    	email TEXT NOT NULL,
    	role TEXT NULL,
		password TEXT NOT NULL,
    	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);`,
		"CREATE_OUTBOX_EVENTS": `CREATE TABLE IF NOT EXISTS outbox_events (
    	id SERIAL PRIMARY KEY,
    	event_type TEXT NOT NULL,
    	payload JSONB NOT NULL,
    	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    	processed BOOLEAN NOT NULL DEFAULT false
	);`,
		"CREATE_INDEX_idx_audit_events_user":        "CREATE INDEX IF NOT EXISTS idx_audit_events_user_id ON audit_events(user_id);",
		"CREATE_INDEX_idx_audit_events_email":       "CREATE INDEX IF NOT EXISTS idx_audit_events_email ON audit_events(email);",
		"CREATE_INDEX_idx_audit_events_event_type":  "CREATE INDEX IF NOT EXISTS idx_audit_events_event_type ON audit_events(event_type);",
		"CREATE_INDEX_idx_audit_events_occurred_at": "CREATE INDEX IF NOT EXISTS idx_audit_events_occurred_at ON audit_events(occurred_at);",
	}

	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	for _, m := range migrations {
		if _, err = conn.Exec(ctx, m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
