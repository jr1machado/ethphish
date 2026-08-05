-- +goose Up
ALTER TABLE groups ADD COLUMN locked BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
-- SQLite does not support DROP COLUMN in older versions; migration is intentionally left empty
