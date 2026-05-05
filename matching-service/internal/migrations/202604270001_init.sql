-- +goose Up
CREATE SCHEMA IF NOT EXISTS matching;

CREATE TABLE IF NOT EXISTS matching.vacancies (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    pvz_id BIGINT NOT NULL,
    status TEXT NOT NULL,
    title TEXT NULL,
    description TEXT NULL,
    shift_date DATE NOT NULL,
    shift_start_at TIME NULL,
    shift_end_at TIME NULL,
    pay_amount BIGINT NOT NULL,
    slots_total INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS matching.responses (
    id BIGSERIAL PRIMARY KEY,
    vacancy_id BIGINT NOT NULL REFERENCES matching.vacancies (id) ON DELETE CASCADE,
    worker_id BIGINT NOT NULL,
    status TEXT NOT NULL,
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
    status TEXT NOT NULL,
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    pvz_name_snapshot TEXT NOT NULL,
    pvz_address_snapshot TEXT NOT NULL,
    shift_date_snapshot DATE NOT NULL,
    pay_amount_snapshot BIGINT NOT NULL,
    shift_start_at_snapshot TIME NULL,
    shift_end_at_snapshot TIME NULL,
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
    target_role TEXT NOT NULL,
    rating INTEGER NOT NULL,
    body TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_matching_reviews_assignment_author UNIQUE (assignment_id, author_user_id)
);

CREATE TABLE IF NOT EXISTS matching.review_replies (
    id BIGSERIAL PRIMARY KEY,
    review_id BIGINT NOT NULL REFERENCES matching.reviews (id) ON DELETE CASCADE,
    author_user_id BIGINT NOT NULL,
    target_user_id BIGINT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_matching_vacancies_owner_id
    ON matching.vacancies (owner_id);

CREATE INDEX IF NOT EXISTS idx_matching_vacancies_pvz_id
    ON matching.vacancies (pvz_id);

CREATE INDEX IF NOT EXISTS idx_matching_vacancies_status
    ON matching.vacancies (status);

CREATE INDEX IF NOT EXISTS idx_matching_vacancies_shift_date
    ON matching.vacancies (shift_date);

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

-- +goose Down
DROP INDEX IF EXISTS idx_matching_review_replies_review_id_created_at;
DROP INDEX IF EXISTS idx_matching_reviews_target_user_id;
DROP INDEX IF EXISTS idx_matching_assignments_status;
DROP INDEX IF EXISTS idx_matching_assignments_worker_id;
DROP INDEX IF EXISTS idx_matching_assignments_owner_id;
DROP INDEX IF EXISTS idx_matching_responses_status;
DROP INDEX IF EXISTS idx_matching_responses_worker_id;
DROP INDEX IF EXISTS idx_matching_responses_vacancy_id;
DROP INDEX IF EXISTS idx_matching_vacancies_shift_date;
DROP INDEX IF EXISTS idx_matching_vacancies_status;
DROP INDEX IF EXISTS idx_matching_vacancies_pvz_id;
DROP INDEX IF EXISTS idx_matching_vacancies_owner_id;
DROP TABLE IF EXISTS matching.review_replies;
DROP TABLE IF EXISTS matching.reviews;
DROP TABLE IF EXISTS matching.assignments;
DROP TABLE IF EXISTS matching.responses;
DROP TABLE IF EXISTS matching.vacancies;
DROP SCHEMA IF EXISTS matching;
