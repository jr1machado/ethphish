-- +goose Up
CREATE TABLE IF NOT EXISTS portal_login_tokens (
    id integer primary key auto_increment,
    tenant_id integer NOT NULL DEFAULT 1,
    email varchar(255) NOT NULL,
    token_hash varchar(255) NOT NULL,
    expires_at datetime NOT NULL,
    used_at datetime,
    created_at datetime
);

CREATE INDEX idx_portal_login_tokens_token_hash ON portal_login_tokens(token_hash);
CREATE INDEX idx_portal_login_tokens_tenant_id ON portal_login_tokens(tenant_id);

-- +goose Down
DROP TABLE portal_login_tokens;
