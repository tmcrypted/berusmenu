-- +goose Up
ALTER TABLE auth.users
    ADD COLUMN IF NOT EXISTS last_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS middle_name TEXT NULL,
    ADD COLUMN IF NOT EXISTS birth_date DATE NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'auth'
          AND table_name = 'users'
          AND column_name = 'full_name'
    ) THEN
        EXECUTE $sql$
            UPDATE auth.users
            SET
                last_name = CASE
                    WHEN last_name = '' THEN COALESCE(NULLIF(split_part(full_name, ' ', 1), ''), '')
                    ELSE last_name
                END,
                first_name = CASE
                    WHEN first_name = '' THEN COALESCE(NULLIF(split_part(full_name, ' ', 2), ''), '')
                    ELSE first_name
                END
        $sql$;
    END IF;
END $$;

ALTER TABLE auth.users
    DROP COLUMN IF EXISTS full_name;

-- +goose Down
ALTER TABLE auth.users
    ADD COLUMN IF NOT EXISTS full_name TEXT NOT NULL DEFAULT '';

UPDATE auth.users
SET full_name = trim(concat_ws(' ', NULLIF(last_name, ''), NULLIF(first_name, ''), NULLIF(middle_name, '')));

ALTER TABLE auth.users
    DROP COLUMN IF EXISTS birth_date,
    DROP COLUMN IF EXISTS middle_name,
    DROP COLUMN IF EXISTS first_name,
    DROP COLUMN IF EXISTS last_name;
