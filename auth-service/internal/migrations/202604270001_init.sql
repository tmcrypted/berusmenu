-- +goose Up
CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    email_verified_at TIMESTAMPTZ NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    full_name TEXT NOT NULL,
    phone TEXT NULL,
    avatar_url TEXT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth.refresh_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES auth.users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ NULL,
    user_agent TEXT NULL,
    ip TEXT NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_refresh_sessions_user_id
    ON auth.refresh_sessions (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_auth_refresh_sessions_user_id;
DROP TABLE IF EXISTS auth.refresh_sessions;
DROP TABLE IF EXISTS auth.users;
DROP SCHEMA IF EXISTS auth;
