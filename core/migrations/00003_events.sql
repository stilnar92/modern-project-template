-- The event log of the core, which is also the outbox (ADR 0003, ADR 0014).
-- The code writes an event and its River job in the same transaction as the
-- action. A plain table for now: ADR 0014 says when to partition it.

-- +goose Up
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

CREATE TABLE events (
    id uuid PRIMARY KEY,
    account_id uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    -- <module>.<thing>_<past tense verb>, docs/dev/events.md
    name text NOT NULL CHECK (name ~ '^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$'),
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    occurred_at timestamptz NOT NULL DEFAULT now(),
    data jsonb NOT NULL DEFAULT '{}'::jsonb
);
-- Reports read the events of one account by time.
CREATE INDEX events_account_occurred_idx ON events (account_id, occurred_at);
ALTER TABLE events ENABLE ROW LEVEL SECURITY;
ALTER TABLE events FORCE ROW LEVEL SECURITY;
CREATE POLICY account_isolation ON events
    USING (account_id = nullif(current_setting('app.account_id', true), '')::uuid)
    WITH CHECK (account_id = nullif(current_setting('app.account_id', true), '')::uuid);

-- +goose Down
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
-- Down undoes only this migration, and roundtrip_test.go runs it. Production
-- never runs Down (backend standard 03), so the drop breaks no client.
-- squawk-ignore ban-drop-table
DROP TABLE events;
