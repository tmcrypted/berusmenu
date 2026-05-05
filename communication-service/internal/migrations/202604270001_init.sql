-- +goose Up
CREATE SCHEMA IF NOT EXISTS communication;

CREATE TABLE IF NOT EXISTS communication.conversations (
    id BIGSERIAL PRIMARY KEY,
    response_id BIGINT NOT NULL UNIQUE,
    vacancy_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    worker_id BIGINT NOT NULL,
    last_message_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS communication.messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES communication.conversations (id) ON DELETE CASCADE,
    sender_user_id BIGINT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS communication.message_reads (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES communication.conversations (id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    last_read_message_id BIGINT NULL REFERENCES communication.messages (id) ON DELETE SET NULL,
    read_at TIMESTAMPTZ NULL,
    CONSTRAINT uq_communication_message_reads_conversation_user UNIQUE (conversation_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_communication_conversations_owner_id
    ON communication.conversations (owner_id);

CREATE INDEX IF NOT EXISTS idx_communication_conversations_worker_id
    ON communication.conversations (worker_id);

CREATE INDEX IF NOT EXISTS idx_communication_conversations_vacancy_id
    ON communication.conversations (vacancy_id);

CREATE INDEX IF NOT EXISTS idx_communication_messages_conversation_id_created_at_id
    ON communication.messages (conversation_id, created_at, id);

-- +goose Down
DROP INDEX IF EXISTS idx_communication_messages_conversation_id_created_at_id;
DROP INDEX IF EXISTS idx_communication_conversations_vacancy_id;
DROP INDEX IF EXISTS idx_communication_conversations_worker_id;
DROP INDEX IF EXISTS idx_communication_conversations_owner_id;
DROP TABLE IF EXISTS communication.message_reads;
DROP TABLE IF EXISTS communication.messages;
DROP TABLE IF EXISTS communication.conversations;
DROP SCHEMA IF EXISTS communication;
