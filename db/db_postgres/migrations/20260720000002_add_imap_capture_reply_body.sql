-- +goose Up
ALTER TABLE imap ADD COLUMN capture_reply_body BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
-- SQLite does not support DROP COLUMN in older versions; migration is intentionally left empty
