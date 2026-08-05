-- +goose Up
-- Compatibility schema for legacy unit fixtures only. PostgreSQL is the
-- supported production datastore.
CREATE TABLE tenants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN NOT NULL DEFAULT 1
);
CREATE TABLE companies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, name),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);
CREATE TABLE tenant_users (
    tenant_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    company_id INTEGER,
    role TEXT NOT NULL DEFAULT 'member',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, user_id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE SET NULL
);
CREATE INDEX idx_companies_tenant_id ON companies(tenant_id);
CREATE INDEX idx_tenant_users_user_id ON tenant_users(user_id);

-- +goose Down
DROP TABLE tenant_users;
DROP TABLE companies;
DROP TABLE tenants;
