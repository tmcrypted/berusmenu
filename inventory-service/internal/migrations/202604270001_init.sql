-- +goose Up
CREATE SCHEMA IF NOT EXISTS inventory;

CREATE TABLE IF NOT EXISTS inventory.pvz (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    status TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    company_name TEXT NOT NULL,
    contact_phone TEXT NOT NULL,
    contact_telegram TEXT NULL,
    description TEXT NULL,
    lat DOUBLE PRECISION NULL,
    lng DOUBLE PRECISION NULL,
    rejection_reason TEXT NULL,
    moderated_by BIGINT NULL,
    moderated_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inventory.pvz_moderation_events (
    id BIGSERIAL PRIMARY KEY,
    pvz_id BIGINT NOT NULL REFERENCES inventory.pvz (id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    actor_user_id BIGINT NOT NULL,
    comment TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_inventory_pvz_owner_id
    ON inventory.pvz (owner_id);

CREATE INDEX IF NOT EXISTS idx_inventory_pvz_status
    ON inventory.pvz (status);

CREATE INDEX IF NOT EXISTS idx_inventory_pvz_moderation_events_pvz_id
    ON inventory.pvz_moderation_events (pvz_id);

-- +goose Down
DROP INDEX IF EXISTS idx_inventory_pvz_moderation_events_pvz_id;
DROP INDEX IF EXISTS idx_inventory_pvz_status;
DROP INDEX IF EXISTS idx_inventory_pvz_owner_id;
DROP TABLE IF EXISTS inventory.pvz_moderation_events;
DROP TABLE IF EXISTS inventory.pvz;
DROP SCHEMA IF EXISTS inventory;
