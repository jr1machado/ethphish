-- +goose Up
-- Normalize legacy SQLite integer flags before the application starts using
-- PostgreSQL's native boolean parameters.
ALTER TABLE sms_logs
    ALTER COLUMN processing TYPE BOOLEAN
    USING CASE WHEN processing IS NULL THEN NULL ELSE processing <> 0 END;

-- +goose Down
ALTER TABLE sms_logs ALTER COLUMN processing TYPE INTEGER
    USING CASE WHEN processing IS NULL THEN NULL WHEN processing THEN 1 ELSE 0 END;
