-- +goose Up
-- SQLite is retained only for legacy test fixtures and import tooling. This
-- compatibility migration keeps those fixtures aligned with the PostgreSQL
-- model; it does not make SQLite a supported production runtime.
INSERT OR IGNORE INTO tenants (id, slug, name, active) VALUES (1, 'default', 'Default tenant', 1);
INSERT OR IGNORE INTO companies (tenant_id, name) VALUES (1, 'Default company');
INSERT OR IGNORE INTO tenant_users (tenant_id, user_id, company_id, role)
SELECT 1, users.id, (SELECT id FROM companies WHERE tenant_id = 1 ORDER BY id LIMIT 1), 'member'
FROM users;

ALTER TABLE campaigns ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE groups ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE templates ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE pages ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE smtp ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE sms_profiles ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE sms_templates ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE imap ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE webhooks ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE reports ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 1;

CREATE INDEX idx_campaigns_tenant_id ON campaigns(tenant_id);
CREATE INDEX idx_groups_tenant_id ON groups(tenant_id);
CREATE INDEX idx_templates_tenant_id ON templates(tenant_id);
CREATE INDEX idx_pages_tenant_id ON pages(tenant_id);
CREATE INDEX idx_smtp_tenant_id ON smtp(tenant_id);
CREATE INDEX idx_sms_profiles_tenant_id ON sms_profiles(tenant_id);
CREATE INDEX idx_sms_templates_tenant_id ON sms_templates(tenant_id);
CREATE INDEX idx_imap_tenant_id ON imap(tenant_id);
CREATE INDEX idx_webhooks_tenant_id ON webhooks(tenant_id);
CREATE INDEX idx_reports_tenant_id ON reports(tenant_id);

-- +goose Down
DROP INDEX idx_reports_tenant_id;
DROP INDEX idx_webhooks_tenant_id;
DROP INDEX idx_imap_tenant_id;
DROP INDEX idx_sms_templates_tenant_id;
DROP INDEX idx_sms_profiles_tenant_id;
DROP INDEX idx_smtp_tenant_id;
DROP INDEX idx_pages_tenant_id;
DROP INDEX idx_templates_tenant_id;
DROP INDEX idx_groups_tenant_id;
DROP INDEX idx_campaigns_tenant_id;
