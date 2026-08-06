-- +goose Up
CREATE TABLE IF NOT EXISTS portal_login_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at DATETIME NOT NULL,
    used_at DATETIME,
    created_at DATETIME
);

CREATE INDEX idx_portal_login_tokens_token_hash ON portal_login_tokens(token_hash);
CREATE INDEX idx_portal_login_tokens_tenant_id ON portal_login_tokens(tenant_id);

-- +goose Down
DROP TABLE portal_login_tokens;
