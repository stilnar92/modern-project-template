#!/bin/sh
# Creates the database and the three roles of ADR 0009. Postgres runs this
# file once, when the data volume is empty. The admin user (postgres) is a
# superuser and is only for administration: superusers ignore RLS.
set -eu

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres \
  --set=owner_pw="$OWNER_DB_PASSWORD" \
  --set=app_pw="$APP_DB_PASSWORD" \
  --set=system_pw="$SYSTEM_DB_PASSWORD" <<'SQL'
-- Schema owner: runs the migrations. FORCE RLS applies to it too.
CREATE ROLE app_owner LOGIN NOSUPERUSER NOBYPASSRLS PASSWORD :'owner_pw';
-- Application: normal work of the api and the worker. RLS applies.
CREATE ROLE app_app LOGIN NOSUPERUSER NOBYPASSRLS PASSWORD :'app_pw';
-- System: only the marked places in ADR 0009. It ignores RLS.
CREATE ROLE app_system LOGIN NOSUPERUSER BYPASSRLS PASSWORD :'system_pw';

-- The owner may switch to the system role only with an explicit SET ROLE,
-- so a migration can hand a SECURITY DEFINER function to it (ADR 0013).
-- INHERIT FALSE: the owner does not get BYPASSRLS by default.
GRANT app_system TO app_owner WITH INHERIT FALSE, SET TRUE;

CREATE DATABASE app OWNER app_owner;
SQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname app <<'SQL'
-- Only the owner creates objects in the public schema.
ALTER SCHEMA public OWNER TO app_owner;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON DATABASE app FROM PUBLIC;
GRANT CONNECT ON DATABASE app TO app_app, app_system;
SQL
