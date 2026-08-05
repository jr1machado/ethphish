-- +goose Up
-- SQLite historically represented this value as an integer. PostgreSQL must
-- store it as a native boolean so GORM can persist the Campaign.HTTPAuth field.
ALTER TABLE campaigns
    ALTER COLUMN http_auth TYPE BOOLEAN
    USING CASE WHEN http_auth IS NULL THEN NULL ELSE http_auth <> 0 END;

-- +goose Down
ALTER TABLE campaigns
    ALTER COLUMN http_auth TYPE INTEGER
    USING CASE WHEN http_auth IS NULL THEN NULL WHEN http_auth THEN 1 ELSE 0 END;
