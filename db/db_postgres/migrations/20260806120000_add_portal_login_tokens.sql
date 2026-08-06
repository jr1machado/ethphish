-- +goose Up
CREATE TABLE IF NOT EXISTS portal_login_tokens (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL DEFAULT 1,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP
);

CREATE INDEX idx_portal_login_tokens_token_hash ON portal_login_tokens(token_hash);
CREATE INDEX idx_portal_login_tokens_tenant_id ON portal_login_tokens(tenant_id);

-- +goose Down
DROP TABLE portal_login_tokens;
