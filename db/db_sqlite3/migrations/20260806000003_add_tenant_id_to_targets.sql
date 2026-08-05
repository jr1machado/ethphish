-- +goose Up
ALTER TABLE targets ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 1;
CREATE INDEX idx_targets_tenant_id ON targets(tenant_id);

-- +goose Down
DROP INDEX idx_targets_tenant_id;
