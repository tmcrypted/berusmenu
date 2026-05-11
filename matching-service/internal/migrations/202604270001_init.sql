-- +goose Up
CREATE SCHEMA IF NOT EXISTS matching;

CREATE TABLE IF NOT EXISTS matching.vacancies (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    pvz_id BIGINT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('published', 'assigned', 'completed', 'cancelled')),
    pvz_name_snapshot TEXT NOT NULL DEFAULT '',
    title TEXT NULL,
    description TEXT NULL,
    shift_start_at TIMESTAMPTZ NOT NULL,
    shift_end_at TIMESTAMPTZ NOT NULL,
    pvz_address_snapshot TEXT NOT NULL DEFAULT '',
    pvz_city_snapshot TEXT NOT NULL DEFAULT '',
    pay_amount BIGINT NOT NULL CHECK (pay_amount > 0),
    slots_total INTEGER NOT NULL DEFAULT 1 CHECK (slots_total = 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_matching_vacancies_shift_order CHECK (shift_end_at > shift_start_at)
);

CREATE TABLE IF NOT EXISTS matching.responses (
    id BIGSERIAL PRIMARY KEY,
    vacancy_id BIGINT NOT NULL REFERENCES matching.vacancies (id) ON DELETE CASCADE,
    worker_id BIGINT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'accepted', 'rejected', 'withdrawn')),
    message TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_matching_responses_vacancy_worker UNIQUE (vacancy_id, worker_id)
);

CREATE TABLE IF NOT EXISTS matching.assignments (
    id BIGSERIAL PRIMARY KEY,
    vacancy_id BIGINT NOT NULL UNIQUE REFERENCES matching.vacancies (id) ON DELETE CASCADE,
    response_id BIGINT NOT NULL UNIQUE REFERENCES matching.responses (id) ON DELETE CASCADE,
    owner_id BIGINT NOT NULL,
    worker_id BIGINT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'completed', 'cancelled')),
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    pvz_name_snapshot TEXT NOT NULL,
    pvz_address_snapshot TEXT NOT NULL,
    pvz_city_snapshot TEXT NOT NULL DEFAULT '',
    shift_start_at_snapshot TIMESTAMPTZ NOT NULL,
    pay_amount_snapshot BIGINT NOT NULL,
    shift_end_at_snapshot TIMESTAMPTZ NOT NULL,
    owner_name_snapshot TEXT NOT NULL,
    worker_name_snapshot TEXT NOT NULL,
    completed_at TIMESTAMPTZ NULL,
    cancelled_at TIMESTAMPTZ NULL,
    cancelled_by_user_id BIGINT NULL,
    cancel_reason TEXT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS matching.reviews (
    id BIGSERIAL PRIMARY KEY,
    assignment_id BIGINT NOT NULL REFERENCES matching.assignments (id) ON DELETE CASCADE,
    author_user_id BIGINT NOT NULL,
    target_user_id BIGINT NOT NULL,
    target_role TEXT NOT NULL CHECK (target_role IN ('owner', 'worker')),
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    body TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_matching_reviews_assignment_author UNIQUE (assignment_id, author_user_id)
);

CREATE TABLE IF NOT EXISTS matching.review_replies (
    id BIGSERIAL PRIMARY KEY,
    review_id BIGINT NOT NULL REFERENCES matching.reviews (id) ON DELETE CASCADE,
    author_user_id BIGINT NOT NULL,
    target_user_id BIGINT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS matching.user_ratings (
    user_id BIGINT PRIMARY KEY,
    avg_rating DOUBLE PRECISION NOT NULL DEFAULT 0,
    reviews_count BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS matching.outbox_events (
    id BIGSERIAL PRIMARY KEY,
    event_id TEXT NOT NULL UNIQUE,
    topic TEXT NOT NULL,
    event_type TEXT NOT NULL,
    event_key TEXT NOT NULL,
    payload BYTEA NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'published')),
    attempts INTEGER NOT NULL DEFAULT 0,
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error TEXT NULL,
    published_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS matching.admin_events (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id BIGINT NOT NULL,
    target_type TEXT NOT NULL,
    target_id BIGINT NOT NULL,
    action TEXT NOT NULL,
    previous_state TEXT NULL,
    new_state TEXT NULL,
    reason TEXT NOT NULL,
    comment TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_matching_vacancies_owner_id
    ON matching.vacancies (owner_id);

CREATE INDEX IF NOT EXISTS idx_matching_vacancies_pvz_id
    ON matching.vacancies (pvz_id);

CREATE INDEX IF NOT EXISTS idx_matching_vacancies_status
    ON matching.vacancies (status);

CREATE INDEX IF NOT EXISTS idx_matching_vacancies_shift_start_at
    ON matching.vacancies (shift_start_at);

CREATE INDEX IF NOT EXISTS idx_matching_responses_vacancy_id
    ON matching.responses (vacancy_id);

CREATE INDEX IF NOT EXISTS idx_matching_responses_worker_id
    ON matching.responses (worker_id);

CREATE INDEX IF NOT EXISTS idx_matching_responses_status
    ON matching.responses (status);

CREATE INDEX IF NOT EXISTS idx_matching_assignments_owner_id
    ON matching.assignments (owner_id);

CREATE INDEX IF NOT EXISTS idx_matching_assignments_worker_id
    ON matching.assignments (worker_id);

CREATE INDEX IF NOT EXISTS idx_matching_assignments_status
    ON matching.assignments (status);

CREATE INDEX IF NOT EXISTS idx_matching_reviews_target_user_id
    ON matching.reviews (target_user_id);

CREATE INDEX IF NOT EXISTS idx_matching_review_replies_review_id_created_at
    ON matching.review_replies (review_id, created_at);

CREATE INDEX IF NOT EXISTS idx_matching_outbox_events_available
    ON matching.outbox_events (available_at, id)
    WHERE status IN ('pending', 'processing');

CREATE INDEX IF NOT EXISTS idx_matching_admin_events_target
    ON matching.admin_events (target_type, target_id, created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_matching_admin_events_target;
DROP INDEX IF EXISTS idx_matching_outbox_events_available;
DROP INDEX IF EXISTS idx_matching_review_replies_review_id_created_at;
DROP INDEX IF EXISTS idx_matching_reviews_target_user_id;
DROP INDEX IF EXISTS idx_matching_assignments_status;
DROP INDEX IF EXISTS idx_matching_assignments_worker_id;
DROP INDEX IF EXISTS idx_matching_assignments_owner_id;
DROP INDEX IF EXISTS idx_matching_responses_status;
DROP INDEX IF EXISTS idx_matching_responses_worker_id;
DROP INDEX IF EXISTS idx_matching_responses_vacancy_id;
DROP INDEX IF EXISTS idx_matching_vacancies_shift_start_at;
DROP INDEX IF EXISTS idx_matching_vacancies_status;
DROP INDEX IF EXISTS idx_matching_vacancies_pvz_id;
DROP INDEX IF EXISTS idx_matching_vacancies_owner_id;
DROP TABLE IF EXISTS matching.admin_events;
DROP TABLE IF EXISTS matching.outbox_events;
DROP TABLE IF EXISTS matching.user_ratings;
DROP TABLE IF EXISTS matching.review_replies;
DROP TABLE IF EXISTS matching.reviews;
DROP TABLE IF EXISTS matching.assignments;
DROP TABLE IF EXISTS matching.responses;
DROP TABLE IF EXISTS matching.vacancies;
DROP SCHEMA IF EXISTS matching;
