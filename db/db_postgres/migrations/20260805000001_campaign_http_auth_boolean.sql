-- +goose Up
-- The PostgreSQL baseline creates http_auth as BOOLEAN. This migration is
-- retained for sequence compatibility with the legacy SQLite history.
SELECT 1;

-- +goose Down
SELECT 1;
