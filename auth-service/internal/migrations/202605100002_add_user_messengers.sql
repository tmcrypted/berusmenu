-- +goose Up
ALTER TABLE auth.users
    ADD COLUMN IF NOT EXISTS telegram_username TEXT NULL,
    ADD COLUMN IF NOT EXISTS max_username TEXT NULL;

-- +goose Down
ALTER TABLE auth.users
    DROP COLUMN IF EXISTS max_username,
    DROP COLUMN IF EXISTS telegram_username;
