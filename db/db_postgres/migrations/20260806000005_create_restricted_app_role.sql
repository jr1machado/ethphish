-- +goose Up
-- The migration runner connects as the privileged/owning role (superuser in
-- the default PostgreSQL image), which always bypasses row level security.
-- The server's runtime connection must use a distinct, non-superuser role
-- with no BYPASSRLS attribute so the tenant_isolation policies created in
-- 20260806000004 are actually enforced. This role only receives DML grants;
-- schema changes stay migration-only.
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'ethphish_app') THEN
        CREATE ROLE ethphish_app LOGIN PASSWORD 'development-only-change-me-app';
    END IF;
END $$;
-- +goose StatementEnd

GRANT CONNECT ON DATABASE ethphish TO ethphish_app;
GRANT USAGE ON SCHEMA public TO ethphish_app;
-- The server's boot path re-runs `CREATE TABLE IF NOT EXISTS goose_db_version`
-- on every start (see migration/runner.go ensureVersionTable); PostgreSQL
-- checks CREATE on the schema before checking whether the table already
-- exists, so this grant is required even though db-migrate already created it.
GRANT CREATE ON SCHEMA public TO ethphish_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ethphish_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ethphish_app;
ALTER DEFAULT PRIVILEGES FOR ROLE ethphish IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ethphish_app;
ALTER DEFAULT PRIVILEGES FOR ROLE ethphish IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO ethphish_app;

-- +goose Down
ALTER DEFAULT PRIVILEGES FOR ROLE ethphish IN SCHEMA public REVOKE USAGE, SELECT ON SEQUENCES FROM ethphish_app;
ALTER DEFAULT PRIVILEGES FOR ROLE ethphish IN SCHEMA public REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM ethphish_app;
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM ethphish_app;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM ethphish_app;
REVOKE CREATE ON SCHEMA public FROM ethphish_app;
REVOKE USAGE ON SCHEMA public FROM ethphish_app;
REVOKE CONNECT ON DATABASE ethphish FROM ethphish_app;
DROP ROLE IF EXISTS ethphish_app;
