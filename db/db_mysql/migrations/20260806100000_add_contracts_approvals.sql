-- +goose Up
CREATE TABLE IF NOT EXISTS contracts (
    id integer primary key auto_increment,
    tenant_id integer NOT NULL DEFAULT 1,
    name varchar(255) NOT NULL,
    client_name varchar(255) NOT NULL,
    status varchar(50) NOT NULL DEFAULT 'draft',
    created_by integer NOT NULL,
    created_at datetime,
    updated_at datetime
);

CREATE TABLE IF NOT EXISTS contract_versions (
    id integer primary key auto_increment,
    contract_id integer NOT NULL,
    version_number integer NOT NULL,
    file_path varchar(500) NOT NULL,
    file_name varchar(255) NOT NULL,
    scope_description text,
    uploaded_by integer NOT NULL,
    uploaded_at datetime,
    notes text
);

CREATE TABLE IF NOT EXISTS contract_approvers (
    id integer primary key auto_increment,
    contract_id integer NOT NULL,
    name varchar(255) NOT NULL,
    email varchar(255) NOT NULL,
    created_at datetime
);

CREATE TABLE IF NOT EXISTS client_users (
    id integer primary key auto_increment,
    tenant_id integer NOT NULL DEFAULT 1,
    email varchar(255) NOT NULL,
    name varchar(255),
    email_verified_at datetime,
    created_at datetime
);

CREATE TABLE IF NOT EXISTS client_sessions (
    id integer primary key auto_increment,
    client_user_id integer NOT NULL,
    session_token_hash varchar(255) NOT NULL,
    expires_at datetime NOT NULL,
    created_at datetime
);

CREATE TABLE IF NOT EXISTS approval_requests (
    id integer primary key auto_increment,
    contract_version_id integer NOT NULL,
    campaign_id integer,
    status varchar(50) NOT NULL DEFAULT 'pending',
    magic_link_token_hash varchar(255) NOT NULL,
    token_expires_at datetime NOT NULL,
    requested_at datetime,
    decided_at datetime,
    decided_by integer,
    last_reminder_sent_at datetime
);

CREATE TABLE IF NOT EXISTS approval_comments (
    id integer primary key auto_increment,
    approval_request_id integer NOT NULL,
    author_type varchar(20) NOT NULL,
    author_id integer NOT NULL,
    body text NOT NULL,
    created_at datetime
);

ALTER TABLE campaigns ADD COLUMN contract_id integer;

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
