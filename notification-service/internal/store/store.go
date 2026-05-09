package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) ReserveEvent(ctx context.Context, eventID string, eventType string, timeout time.Duration) (bool, error) {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	sql := `
INSERT INTO notification.processed_events (event_id, event_type, status, reserved_at)
VALUES ($1, $2, 'processing', NOW())
ON CONFLICT (event_id) DO UPDATE
SET status = 'processing',
    reserved_at = NOW(),
    event_type = EXCLUDED.event_type
WHERE notification.processed_events.status = 'processing'
  AND notification.processed_events.reserved_at < NOW() - ($3 * INTERVAL '1 second')
RETURNING event_id`

	var reserved string
	err := s.pool.QueryRow(ctx, sql, eventID, eventType, int64(timeout.Seconds())).Scan(&reserved)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Store) MarkEventProcessed(ctx context.Context, eventID string) error {
	_, err := s.pool.Exec(ctx, `
UPDATE notification.processed_events
SET status = 'processed', processed_at = NOW()
WHERE event_id = $1`, eventID)
	return err
}

func (s *Store) ReleaseEvent(ctx context.Context, eventID string) error {
	_, err := s.pool.Exec(ctx, `
DELETE FROM notification.processed_events
WHERE event_id = $1 AND status = 'processing'`, eventID)
	return err
}
