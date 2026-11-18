package cursor

import (
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/stilnar92/modern-project-template/core/internal/platform/testdb"
)

func TestRoundTrip(t *testing.T) {
	k := Key{CreatedAt: time.Date(2026, 11, 12, 10, 0, 0, 123456000, time.UTC), ID: uuid.NewV7()}
	s := Encode(k)
	if len(s) != 32 {
		t.Fatalf("cursor %q has %d characters, want 32", s, len(s))
	}
	got, ok, err := Decode(s)
	if err != nil || !ok || !got.CreatedAt.Equal(k.CreatedAt) || got.ID != k.ID {
		t.Fatalf("got %+v, %v, %v; want %+v", got, ok, err, k)
	}
}

func TestDecodeRejects(t *testing.T) {
	for _, s := range []string{"not base64!", "AAAA", Encode(Key{}) + "AA"} {
		if _, _, err := Decode(s); !errors.Is(err, ErrInvalid) {
			t.Errorf("Decode(%q): got %v, want ErrInvalid", s, err)
		}
	}
	if _, ok, err := Decode(""); ok || err != nil {
		t.Fatalf("empty cursor: ok %v, err %v; want the first page", ok, err)
	}
}

// A client sends any text as a cursor. Decode must never panic, and a
// cursor that it accepts must encode back to the same text.
func FuzzDecode(f *testing.F) {
	f.Add("")
	f.Add(Encode(Key{CreatedAt: time.Now(), ID: uuid.NewV7()}))
	f.Add("%%%")
	f.Fuzz(func(t *testing.T, s string) {
		k, ok, err := Decode(s)
		if err != nil || !ok {
			return
		}
		if Encode(k) != s {
			t.Fatalf("Encode(Decode(%q)) = %q", s, Encode(k))
		}
	})
}

func TestLimit(t *testing.T) {
	n := func(v int) *int { return &v }
	for _, tt := range []struct {
		in   *int
		want int
	}{{nil, DefaultLimit}, {n(0), DefaultLimit}, {n(-5), DefaultLimit}, {n(7), 7}, {n(1000), MaxLimit}} {
		if got := Limit(tt.in); got != tt.want {
			t.Errorf("Limit(%v) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

// Pages through a real table with the query from the package comment. Rows
// with the same created_at check the tie-break on id. Each row comes once,
// in order, and the last page has no cursor.
func TestPagesThroughPostgres(t *testing.T) {
	d := testdb.Start(t)
	ctx := t.Context()
	conn, err := d.Owner.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, `CREATE TEMP TABLE things (id uuid PRIMARY KEY, created_at timestamptz NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	same := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	const total = 23
	for i := range total {
		at := same
		if i%3 == 0 {
			at = same.Add(time.Duration(i) * time.Second)
		}
		if _, err := conn.Exec(ctx, `INSERT INTO things VALUES ($1, $2)`, uuid.NewV7(), at); err != nil {
			t.Fatal(err)
		}
	}

	const query = `SELECT id, created_at FROM things
WHERE ($1::timestamptz IS NULL OR (created_at, id) < ($1, $2::uuid))
ORDER BY created_at DESC, id DESC LIMIT $3`
	seen := map[uuid.UUID]bool{}
	var last *Key
	next, pages := "", 0
	for {
		k, ok, err := Decode(next)
		if err != nil {
			t.Fatal(err)
		}
		var after *time.Time
		var afterID *uuid.UUID
		if ok {
			after, afterID = &k.CreatedAt, &k.ID
		}
		limit := Limit(new(10))
		rows, err := conn.Query(ctx, query, after, afterID, limit+1)
		if err != nil {
			t.Fatal(err)
		}
		keys, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (Key, error) {
			var k Key
			err := r.Scan(&k.ID, &k.CreatedAt)
			return k, err
		})
		if err != nil {
			t.Fatal(err)
		}
		var page []Key
		page, next = Page(keys, limit, func(k Key) Key { return k })
		pages++
		for _, k := range page {
			if seen[k.ID] {
				t.Fatalf("row %s comes twice", k.ID)
			}
			seen[k.ID] = true
			if last != nil && (k.CreatedAt.After(last.CreatedAt) || (k.CreatedAt.Equal(last.CreatedAt) && k.ID.Compare(last.ID) >= 0)) {
				t.Fatalf("row %v comes after %v: wrong order", k, *last)
			}
			last = &k
		}
		if next == "" {
			break
		}
	}
	if len(seen) != total || pages != 3 {
		t.Fatalf("saw %d rows in %d pages, want %d rows in 3 pages", len(seen), pages, total)
	}
}
