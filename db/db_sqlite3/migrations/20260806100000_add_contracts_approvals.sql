-- +goose Up
CREATE TABLE IF NOT EXISTS contracts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1,
    name VARCHAR(255) NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    created_by INTEGER NOT NULL,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS contract_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    contract_id INTEGER NOT NULL,
    version_number INTEGER NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    scope_description TEXT,
    uploaded_by INTEGER NOT NULL,
    uploaded_at DATETIME,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS contract_approvers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    contract_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at DATETIME
);

CREATE TABLE IF NOT EXISTS client_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    email_verified_at DATETIME,
    created_at DATETIME
);

CREATE TABLE IF NOT EXISTS client_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    client_user_id INTEGER NOT NULL,
    session_token_hash VARCHAR(255) NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME
);

CREATE TABLE IF NOT EXISTS approval_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    contract_version_id INTEGER NOT NULL,
    campaign_id INTEGER,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    magic_link_token_hash VARCHAR(255) NOT NULL,
    token_expires_at DATETIME NOT NULL,
    requested_at DATETIME,
    decided_at DATETIME,
    decided_by INTEGER,
    last_reminder_sent_at DATETIME
);

CREATE TABLE IF NOT EXISTS approval_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    approval_request_id INTEGER NOT NULL,
    author_type VARCHAR(20) NOT NULL,
    author_id INTEGER NOT NULL,
    body TEXT NOT NULL,
    created_at DATETIME
);

ALTER TABLE campaigns ADD COLUMN contract_id INTEGER;

CREATE INDEX idx_contracts_tenant_id ON contracts(tenant_id);
CREATE INDEX idx_contract_versions_contract_id ON contract_versions(contract_id);
CREATE INDEX idx_contract_approvers_contract_id ON contract_approvers(contract_id);
CREATE INDEX idx_client_users_email ON client_users(email);
CREATE INDEX idx_client_sessions_client_user_id ON client_sessions(client_user_id);
CREATE INDEX idx_client_sessions_token_hash ON client_sessions(session_token_hash);
CREATE INDEX idx_approval_requests_contract_version_id ON approval_requests(contract_version_id);
CREATE INDEX idx_approval_requests_status ON approval_requests(status);
CREATE INDEX idx_approval_requests_token_hash ON approval_requests(magic_link_token_hash);
CREATE INDEX idx_approval_comments_approval_request_id ON approval_comments(approval_request_id);
CREATE INDEX idx_campaigns_contract_id ON campaigns(contract_id);

-- +goose Down
DROP TABLE approval_comments;
DROP TABLE approval_requests;
DROP TABLE client_sessions;
DROP TABLE client_users;
DROP TABLE contract_approvers;
DROP TABLE contract_versions;
DROP TABLE contracts;
ALTER TABLE campaigns DROP COLUMN contract_id;
