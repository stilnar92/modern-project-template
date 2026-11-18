-- Accounts and server sessions (ADR 0012, ADR 0013).
-- accounts is minimal: the product adds its columns later (expand).
-- A session keeps only the SHA-256 hash of its token, never the token.

-- +goose Up
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

CREATE TABLE accounts (
    id uuid PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now()
);
-- The row of an account is visible only inside its own account context.
ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE accounts FORCE ROW LEVEL SECURITY;
CREATE POLICY account_isolation ON accounts
    USING (id = nullif(current_setting('app.account_id', true), '')::uuid)
    WITH CHECK (id = nullif(current_setting('app.account_id', true), '')::uuid);

CREATE TABLE sessions (
    token_hash bytea PRIMARY KEY CHECK (octet_length(token_hash) = 32),
    account_id uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);
-- "End all sessions of an account" and the foreign key use this index.
CREATE INDEX sessions_account_id_idx ON sessions (account_id);
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions FORCE ROW LEVEL SECURITY;
CREATE POLICY account_isolation ON sessions
    USING (account_id = nullif(current_setting('app.account_id', true), '')::uuid)
    WITH CHECK (account_id = nullif(current_setting('app.account_id', true), '')::uuid);

-- The session lookup runs before the account is known, so RLS cannot apply
-- yet. This function is the only way around RLS for sessions (ADR 0013): it
-- returns only the account of one live session, or NULL. The system role
-- owns it, so it reads the table without RLS. The app role can only call it.
CREATE FUNCTION session_account(p_token_hash bytea) RETURNS uuid
    LANGUAGE sql STABLE SECURITY DEFINER
    SET search_path = public, pg_temp
AS $$
    SELECT s.account_id FROM sessions s
    WHERE s.token_hash = p_token_hash AND s.expires_at > now()
$$;
-- Set the rights first: only the owner can change them, and the hand-over
-- keeps them. After the hand-over this role is no longer the owner, and a
-- REVOKE would only give a warning.
REVOKE ALL ON FUNCTION session_account(bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION session_account(bytea) TO app_app;
-- A new owner needs CREATE on the schema. The system role gets it only for
-- the hand-over, in this transaction.
GRANT CREATE ON SCHEMA public TO app_system;
ALTER FUNCTION session_account(bytea) OWNER TO app_system;
REVOKE CREATE ON SCHEMA public FROM app_system;

-- +goose Down
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
-- Down undoes only this migration, and roundtrip_test.go runs it. Production
-- never runs Down (backend standard 03), so the drops break no client.
-- squawk-ignore ban-drop-function
DROP FUNCTION session_account(bytea);
-- squawk-ignore ban-drop-table
DROP TABLE sessions;
-- squawk-ignore ban-drop-table
DROP TABLE accounts;
