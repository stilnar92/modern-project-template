-- name: InsertEvent :exec
INSERT INTO events (id, account_id, name, version, occurred_at, data)
VALUES (@id, @account_id, @name, @version, @occurred_at, @data);

-- name: GetEvent :one
SELECT id, account_id, name, version, occurred_at, data FROM events WHERE id = @id;
