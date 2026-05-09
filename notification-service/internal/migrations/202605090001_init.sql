-- +goose Up
CREATE SCHEMA IF NOT EXISTS notification;

CREATE TABLE IF NOT EXISTS notification.processed_events (
    event_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('processing', 'processed')),
    reserved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_notification_processed_events_status_reserved
    ON notification.processed_events (status, reserved_at);

-- +goose Down
DROP INDEX IF EXISTS idx_notification_processed_events_status_reserved;
DROP TABLE IF EXISTS notification.processed_events;
DROP SCHEMA IF EXISTS notification;
