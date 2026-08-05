-- +goose Up
-- Add MFA fields to pages table
ALTER TABLE pages ADD COLUMN enable_mfa BOOLEAN DEFAULT FALSE;
ALTER TABLE pages ADD COLUMN mfa_sms_profile_id INTEGER DEFAULT 0;
ALTER TABLE pages ADD COLUMN mfa_from VARCHAR(255) DEFAULT '';
ALTER TABLE pages ADD COLUMN mfa_message TEXT DEFAULT '';
ALTER TABLE pages ADD COLUMN mfa_code_length INTEGER DEFAULT 6;
ALTER TABLE pages ADD COLUMN mfa_code_type VARCHAR(20) DEFAULT 'numeric';
ALTER TABLE pages ADD COLUMN mfa_inject_page BOOLEAN DEFAULT TRUE;
ALTER TABLE pages ADD COLUMN mfa_page_html TEXT DEFAULT '';

-- Create mfa_codes table for storing generated MFA codes
CREATE TABLE IF NOT EXISTS mfa_codes (
    id BIGSERIAL PRIMARY KEY,
    r_id VARCHAR(255) NOT NULL,
    code VARCHAR(20) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    verified BOOLEAN DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_mfa_codes_rid ON mfa_codes(r_id);

-- +goose Down
-- SQLite doesn't support DROP COLUMN directly
-- For a proper down migration, the table would need to be recreated
DROP TABLE IF EXISTS mfa_codes;
