-- +goose Up
ALTER TABLE auth.users
    ADD COLUMN IF NOT EXISTS blocked_until TIMESTAMPTZ NULL;

-- +goose Down
ALTER TABLE auth.users
    DROP COLUMN IF EXISTS blocked_until;
