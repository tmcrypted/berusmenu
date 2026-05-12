ALTER TABLE matching.vacancies
    ADD COLUMN IF NOT EXISTS pvz_city_snapshot TEXT NOT NULL DEFAULT '';
