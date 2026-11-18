// Package cursor is the keyset pagination of list endpoints (backend
// standard 03, ADR 0012). A list is ordered by (created_at, id), newest
// first. The cursor is the key of the last row of a page, so the next page
// starts after it: no OFFSET, and a new row does not shift the pages.
//
// The query of a list endpoint:
//
//	-- name: ListThings :many
//	SELECT ... FROM things
//	WHERE account_id = @account_id
//	  AND (sqlc.narg(after_created_at)::timestamptz IS NULL
//	       OR (created_at, id) < (sqlc.narg(after_created_at), sqlc.narg(after_id)::uuid))
//	ORDER BY created_at DESC, id DESC
//	LIMIT @row_limit;
//
// The index starts with account_id: (account_id, created_at DESC, id DESC).
// The handler reads ?cursor= and ?limit=, asks for Limit(n)+1 rows, and
// calls Page. The response has the items and next_cursor (empty on the last page).
package cursor

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"time"
	"uuid"
)

// Limits of the page size.
const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// ErrInvalid means that the client sent a cursor that this server did not
// make. The handler answers 400.
var ErrInvalid = errors.New("invalid cursor")

// Key is the position of a row in a list.
type Key struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

// size: 8 bytes of microseconds (the precision of timestamptz) and 16 bytes of the ID.
const size = 8 + 16

// Encode makes the opaque cursor text: base64url, 32 characters. Clients
// must not read it: its form can change.
func Encode(k Key) string {
	var b [size]byte
	binary.BigEndian.PutUint64(b[:8], uint64(k.CreatedAt.UnixMicro())) //nolint:gosec // negative times only before 1970, round trip keeps them
	copy(b[8:], k.ID[:])
	return base64.RawURLEncoding.EncodeToString(b[:])
}

// Decode reads a cursor. An empty text means the first page: ok is false.
func Decode(s string) (k Key, ok bool, err error) {
	if s == "" {
		return Key{}, false, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil || len(b) != size {
		return Key{}, false, ErrInvalid
	}
	k.CreatedAt = time.UnixMicro(int64(binary.BigEndian.Uint64(b[:8]))).UTC() //nolint:gosec // the inverse of Encode
	copy(k.ID[:], b[8:])
	// The decoder skips line breaks, so other texts can give the same key.
	// Accept only the text that Encode makes: one key, one cursor.
	if Encode(k) != s {
		return Key{}, false, ErrInvalid
	}
	return k, true, nil
}

// Limit returns the page size for the ?limit= value: the default when it
// is missing or not positive, and at most MaxLimit.
func Limit(n *int) int {
	switch {
	case n == nil || *n <= 0:
		return DefaultLimit
	case *n > MaxLimit:
		return MaxLimit
	default:
		return *n
	}
}

// Page cuts rows to limit and returns the next cursor. The query must ask
// for limit+1 rows: one extra row tells that a next page exists. The cursor
// is empty on the last page.
func Page[T any](rows []T, limit int, key func(T) Key) ([]T, string) {
	if len(rows) <= limit {
		return rows, ""
	}
	rows = rows[:limit]
	return rows, Encode(key(rows[limit-1]))
}
