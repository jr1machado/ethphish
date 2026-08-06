-- +goose Up
-- Enables PostgreSQL RLS as a backstop on every table that now carries
-- tenant_id. Rows are visible/writable only when the session's
-- ethphish.tenant_id matches, or when no tenant_id has been set for the
-- session at all. The NULL passthrough keeps legacy background workers
-- (the IMAP monitor's outer scan, the report queue drainer, scheduled
-- cleanup) working unchanged: they never call set_config and so keep the
-- cross-tenant visibility they already require by design. Every
-- request-driven code path in this codebase now routes tenant-owned reads
-- and writes through withTenantTransaction, which always sets
-- ethphish.tenant_id first, so RLS is fully enforced there.
--
-- +goose StatementBegin
DO $$
DECLARE
    t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY[
        'campaigns', 'groups', 'targets', 'templates', 'pages',
        'smtp', 'sms_profiles', 'sms_templates', 'imap', 'webhooks', 'reports'
    ] LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        -- NULLIF(...,'')::bigint never raises: once a session has called
        -- set_config for this custom GUC even once, PostgreSQL keeps a
        -- session-local placeholder whose "unset" value reads back as ''
        -- rather than NULL, and casting '' straight to bigint would error.
        EXECUTE format(
            'CREATE POLICY tenant_isolation ON %I ' ||
            'USING (NULLIF(current_setting(''ethphish.tenant_id'', true), '''') IS NULL ' ||
            'OR tenant_id = NULLIF(current_setting(''ethphish.tenant_id'', true), '''')::bigint) ' ||
            'WITH CHECK (NULLIF(current_setting(''ethphish.tenant_id'', true), '''') IS NULL ' ||
            'OR tenant_id = NULLIF(current_setting(''ethphish.tenant_id'', true), '''')::bigint)',
            t
        );
    END LOOP;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
    t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY[
        'campaigns', 'groups', 'targets', 'templates', 'pages',
        'smtp', 'sms_profiles', 'sms_templates', 'imap', 'webhooks', 'reports'
    ] LOOP
        EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
        EXECUTE format('ALTER TABLE %I NO FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I DISABLE ROW LEVEL SECURITY', t);
    END LOOP;
END $$;
-- +goose StatementEnd
