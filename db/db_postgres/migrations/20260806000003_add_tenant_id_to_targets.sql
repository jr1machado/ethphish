-- +goose Up
ALTER TABLE targets ADD COLUMN tenant_id BIGINT;
UPDATE targets SET tenant_id = 1 WHERE tenant_id IS NULL;
ALTER TABLE targets ALTER COLUMN tenant_id SET DEFAULT 1;
ALTER TABLE targets ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE targets ADD CONSTRAINT targets_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);
CREATE INDEX idx_targets_tenant_id ON targets(tenant_id);

-- +goose Down
DROP INDEX idx_targets_tenant_id;
ALTER TABLE targets DROP COLUMN tenant_id;
