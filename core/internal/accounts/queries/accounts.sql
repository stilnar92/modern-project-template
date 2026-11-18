-- name: CreateAccount :exec
INSERT INTO accounts (id) VALUES (@id);

-- name: CreateSession :exec
INSERT INTO sessions (token_hash, account_id, expires_at)
VALUES (@token_hash, @account_id, @expires_at);

-- name: SessionAccount :one
-- The only lookup without an account context (ADR 0013).
-- No live session gives no row, so the code gets pgx.ErrNoRows.
SELECT s.account_id FROM (SELECT session_account(@token_hash::bytea) AS account_id) AS s
WHERE s.account_id IS NOT NULL;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = @token_hash;

-- name: DeleteAccountSessions :exec
DELETE FROM sessions WHERE account_id = @account_id;

-- name: DeleteExpiredSessions :execrows
-- Runs with the system role, for all accounts (ADR 0009: scheduled deletion).
DELETE FROM sessions WHERE expires_at < now();
