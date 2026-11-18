-- Rights of the application and system roles (ADR 0009). The roles exist
-- before the first migration: deploy/postgres/init creates them locally.
-- The schema owner runs the migrations, so the default privileges apply to
-- each table that the owner creates later.

-- +goose Up
-- Each migration sets time limits, so a migration that waits for a lock
-- fails fast and does not block the application (squawk). SET LOCAL keeps
-- them inside the transaction of this migration.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
GRANT USAGE ON SCHEMA public TO app_app, app_system;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_app, app_system;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO app_app, app_system;

-- +goose Down
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    REVOKE USAGE, SELECT ON SEQUENCES FROM app_app, app_system;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM app_app, app_system;
REVOKE USAGE ON SCHEMA public FROM app_app, app_system;
