-- +goose Up
CREATE SCHEMA IF NOT EXISTS matching;

CREATE TABLE IF NOT EXISTS matching.vacancies (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    pvz_id BIGINT NOT NULL,
    status TEXT NOT NULL,
    pvz_name_snapshot TEXT NOT NULL DEFAULT '',
    title TEXT NULL,
    description TEXT NULL,
    shift_start_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    shift_end_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '8 hours',
    pvz_address_snapshot TEXT NOT NULL DEFAULT '',
    pvz_city_snapshot TEXT NOT NULL DEFAULT '',
    pay_amount BIGINT NOT NULL DEFAULT 1,
    slots_total INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE matching.vacancies
    ADD COLUMN IF NOT EXISTS pvz_name_snapshot TEXT NOT NULL DEFAULT '';

ALTER TABLE matching.vacancies
    ADD COLUMN IF NOT EXISTS pvz_address_snapshot TEXT NOT NULL DEFAULT '';

ALTER TABLE matching.vacancies
    ADD COLUMN IF NOT EXISTS pvz_city_snapshot TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'matching'
          AND table_name = 'vacancies'
          AND column_name = 'shift_start_at'
          AND data_type = 'time without time zone'
    ) THEN
        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'matching'
              AND table_name = 'vacancies'
              AND column_name = 'shift_date'
        ) THEN
            EXECUTE $sql$
                ALTER TABLE matching.vacancies
                    ALTER COLUMN shift_start_at TYPE TIMESTAMPTZ
                    USING ((COALESCE(shift_date, CURRENT_DATE)::timestamp + shift_start_at) AT TIME ZONE 'UTC')
            $sql$;
        ELSE
            EXECUTE $sql$
                ALTER TABLE matching.vacancies
                    ALTER COLUMN shift_start_at TYPE TIMESTAMPTZ
                    USING ((CURRENT_DATE::timestamp + shift_start_at) AT TIME ZONE 'UTC')
            $sql$;
        END IF;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'matching'
          AND table_name = 'vacancies'
          AND column_name = 'shift_end_at'
          AND data_type = 'time without time zone'
    ) THEN
        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'matching'
              AND table_name = 'vacancies'
              AND column_name = 'shift_date'
        ) THEN
            EXECUTE $sql$
                ALTER TABLE matching.vacancies
                    ALTER COLUMN shift_end_at TYPE TIMESTAMPTZ
                    USING ((COALESCE(shift_date, CURRENT_DATE)::timestamp + shift_end_at) AT TIME ZONE 'UTC')
            $sql$;
        ELSE
            EXECUTE $sql$
                ALTER TABLE matching.vacancies
                    ALTER COLUMN shift_end_at TYPE TIMESTAMPTZ
                    USING ((CURRENT_DATE::timestamp + shift_end_at) AT TIME ZONE 'UTC')
            $sql$;
        END IF;
    END IF;
END $$;

ALTER TABLE matching.vacancies
    ADD COLUMN IF NOT EXISTS shift_start_at TIMESTAMPTZ;

ALTER TABLE matching.vacancies
    ADD COLUMN IF NOT EXISTS shift_end_at TIMESTAMPTZ;

UPDATE matching.vacancies
SET shift_start_at = COALESCE(shift_start_at, created_at, NOW()),
    shift_end_at = COALESCE(shift_end_at, COALESCE(shift_start_at, created_at, NOW()) + INTERVAL '8 hours'),
    pay_amount = CASE WHEN pay_amount IS NULL OR pay_amount <= 0 THEN 1 ELSE pay_amount END,
    slots_total = 1,
    pvz_name_snapshot = COALESCE(pvz_name_snapshot, ''),
    pvz_address_snapshot = COALESCE(pvz_address_snapshot, ''),
    pvz_city_snapshot = COALESCE(pvz_city_snapshot, '');

UPDATE matching.vacancies
SET shift_end_at = shift_start_at + INTERVAL '8 hours'
WHERE shift_end_at <= shift_start_at;

ALTER TABLE matching.vacancies
    ALTER COLUMN pvz_name_snapshot SET NOT NULL,
    ALTER COLUMN pvz_address_snapshot SET NOT NULL,
    ALTER COLUMN pvz_city_snapshot SET NOT NULL,
    ALTER COLUMN shift_start_at SET NOT NULL,
    ALTER COLUMN shift_end_at SET NOT NULL,
    ALTER COLUMN pay_amount SET NOT NULL,
    ALTER COLUMN slots_total SET NOT NULL,
    ALTER COLUMN slots_total SET DEFAULT 1;

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
    status TEXT NOT NULL DEFAULT 'pending',
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

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'matching' AND table_name = 'reviews') THEN
        UPDATE matching.reviews
        SET rating = LEAST(5, GREATEST(1, rating))
        WHERE rating NOT BETWEEN 1 AND 5;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_matching_vacancies_pay_positive') THEN
        ALTER TABLE matching.vacancies
            ADD CONSTRAINT ck_matching_vacancies_pay_positive CHECK (pay_amount > 0) NOT VALID;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_matching_vacancies_slots_one') THEN
        ALTER TABLE matching.vacancies
            ADD CONSTRAINT ck_matching_vacancies_slots_one CHECK (slots_total = 1) NOT VALID;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_matching_vacancies_shift_order') THEN
        ALTER TABLE matching.vacancies
            ADD CONSTRAINT ck_matching_vacancies_shift_order CHECK (shift_end_at > shift_start_at) NOT VALID;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'matching' AND table_name = 'reviews')
       AND NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_matching_reviews_rating_scale') THEN
        ALTER TABLE matching.reviews
            ADD CONSTRAINT ck_matching_reviews_rating_scale CHECK (rating BETWEEN 1 AND 5) NOT VALID;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_matching_review_replies_review_id_created_at
    ON matching.review_replies (review_id, created_at);

CREATE INDEX IF NOT EXISTS idx_matching_admin_events_target
    ON matching.admin_events (target_type, target_id, created_at);

-- +goose Down
