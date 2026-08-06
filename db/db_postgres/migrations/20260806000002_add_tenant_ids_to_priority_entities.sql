-- +goose Up
-- Every installation upgraded from the single-tenant model receives one
-- explicit legacy tenant. New entities retain that tenant by default until a
-- scoped request explicitly supplies another authorized tenant.
INSERT INTO tenants (id, slug, name, active)
VALUES (1, 'default', 'Default tenant', TRUE)
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('tenants', 'id'), (SELECT MAX(id) FROM tenants));

INSERT INTO companies (tenant_id, name)
SELECT 1, 'Default company'
WHERE NOT EXISTS (SELECT 1 FROM companies WHERE tenant_id = 1);

INSERT INTO tenant_users (tenant_id, user_id, company_id, role)
SELECT 1, users.id, (SELECT id FROM companies WHERE tenant_id = 1 ORDER BY id LIMIT 1), 'member'
FROM users
ON CONFLICT (tenant_id, user_id) DO NOTHING;

ALTER TABLE campaigns ADD COLUMN tenant_id BIGINT;
ALTER TABLE groups ADD COLUMN tenant_id BIGINT;
ALTER TABLE templates ADD COLUMN tenant_id BIGINT;
ALTER TABLE pages ADD COLUMN tenant_id BIGINT;
ALTER TABLE smtp ADD COLUMN tenant_id BIGINT;
ALTER TABLE sms_profiles ADD COLUMN tenant_id BIGINT;
ALTER TABLE sms_templates ADD COLUMN tenant_id BIGINT;
ALTER TABLE imap ADD COLUMN tenant_id BIGINT;
ALTER TABLE webhooks ADD COLUMN tenant_id BIGINT;
ALTER TABLE reports ADD COLUMN tenant_id BIGINT;

UPDATE campaigns SET tenant_id = 1 WHERE tenant_id IS NULL;
UPDATE groups SET tenant_id = 1 WHERE tenant_id IS NULL;
UPDATE templates SET tenant_id = 1 WHERE tenant_id IS NULL;
UPDATE pages SET tenant_id = 1 WHERE tenant_id IS NULL;
UPDATE smtp SET tenant_id = 1 WHERE tenant_id IS NULL;
UPDATE sms_profiles SET tenant_id = 1 WHERE tenant_id IS NULL;
UPDATE sms_templates SET tenant_id = 1 WHERE tenant_id IS NULL;
UPDATE imap SET tenant_id = 1 WHERE tenant_id IS NULL;
UPDATE webhooks SET tenant_id = 1 WHERE tenant_id IS NULL;
UPDATE reports SET tenant_id = 1 WHERE tenant_id IS NULL;

ALTER TABLE campaigns ALTER COLUMN tenant_id SET DEFAULT 1;
ALTER TABLE groups ALTER COLUMN tenant_id SET DEFAULT 1;
ALTER TABLE templates ALTER COLUMN tenant_id SET DEFAULT 1;
ALTER TABLE pages ALTER COLUMN tenant_id SET DEFAULT 1;
ALTER TABLE smtp ALTER COLUMN tenant_id SET DEFAULT 1;
ALTER TABLE sms_profiles ALTER COLUMN tenant_id SET DEFAULT 1;
ALTER TABLE sms_templates ALTER COLUMN tenant_id SET DEFAULT 1;
ALTER TABLE imap ALTER COLUMN tenant_id SET DEFAULT 1;
ALTER TABLE webhooks ALTER COLUMN tenant_id SET DEFAULT 1;
ALTER TABLE reports ALTER COLUMN tenant_id SET DEFAULT 1;

ALTER TABLE campaigns ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE groups ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE templates ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE pages ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE smtp ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE sms_profiles ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE sms_templates ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE imap ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE webhooks ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE reports ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE campaigns ADD CONSTRAINT campaigns_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);
ALTER TABLE groups ADD CONSTRAINT groups_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);
ALTER TABLE templates ADD CONSTRAINT templates_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);
ALTER TABLE pages ADD CONSTRAINT pages_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);
ALTER TABLE smtp ADD CONSTRAINT smtp_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);
ALTER TABLE sms_profiles ADD CONSTRAINT sms_profiles_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);
ALTER TABLE sms_templates ADD CONSTRAINT sms_templates_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);
ALTER TABLE imap ADD CONSTRAINT imap_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);
ALTER TABLE webhooks ADD CONSTRAINT webhooks_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);
ALTER TABLE reports ADD CONSTRAINT reports_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

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
ALTER TABLE reports DROP COLUMN tenant_id;
ALTER TABLE webhooks DROP COLUMN tenant_id;
ALTER TABLE imap DROP COLUMN tenant_id;
ALTER TABLE sms_templates DROP COLUMN tenant_id;
ALTER TABLE sms_profiles DROP COLUMN tenant_id;
ALTER TABLE smtp DROP COLUMN tenant_id;
ALTER TABLE pages DROP COLUMN tenant_id;
ALTER TABLE templates DROP COLUMN tenant_id;
ALTER TABLE groups DROP COLUMN tenant_id;
ALTER TABLE campaigns DROP COLUMN tenant_id;
