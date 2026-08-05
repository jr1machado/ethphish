-- +goose Up
-- The PostgreSQL baseline creates processing as BOOLEAN. This migration is
-- retained for sequence compatibility with the legacy SQLite history.
SELECT 1;

-- +goose Down
SELECT 1;
