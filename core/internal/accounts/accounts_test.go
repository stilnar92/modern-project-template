package accounts

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/stilnar92/modern-project-template/core/internal/platform/db"
	"github.com/stilnar92/modern-project-template/core/internal/platform/testdb"
)

func newService(t *testing.T) (*Service, *testdb.DB) {
	t.Helper()
	d := testdb.Start(t)
	return New(d.App, Options{TTL: time.Hour, SecureCookie: true}), d
}

func TestSessionsIsolation(t *testing.T) {
	_, d := newService(t)
	testdb.AssertIsolation(t, d, testdb.IsolationCase{
		Table: "sessions",
		Insert: func(ctx context.Context, tx pgx.Tx, accountID string) error {
			if _, err := tx.Exec(ctx, "INSERT INTO accounts (id) VALUES ($1) ON CONFLICT DO NOTHING", accountID); err != nil {
				return err
			}
			hash := make([]byte, 32)
			_, _ = rand.Read(hash)
			_, err := tx.Exec(ctx, "INSERT INTO sessions (token_hash, account_id, expires_at) VALUES ($1, $2, now() + interval '1 hour')", hash, accountID)
			return err
		},
	})
}

// An account sees only its own row in accounts.
func TestAccountsIsolation(t *testing.T) {
	s, d := newService(t)
	ctx := t.Context()
	a, err := s.CreateAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreateAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = db.WithAccount(ctx, d.App, b.String(), func(tx pgx.Tx) error {
		var n int
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM accounts").Scan(&n); err != nil {
			return err
		}
		if n != 1 {
			t.Errorf("account B sees %d accounts, want 1", n)
		}
		tag, err := tx.Exec(ctx, "DELETE FROM accounts WHERE id = $1", a)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 0 {
			t.Error("account B deletes account A")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	s, d := newService(t)
	ctx := t.Context()
	acc, err := s.CreateAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if acc.String()[14] != '7' {
		t.Fatalf("the account ID is not a UUIDv7: %s", acc)
	}

	sess, err := s.StartSession(ctx, acc)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.SessionAccount(ctx, sess.Token)
	if err != nil || got != acc {
		t.Fatalf("lookup: %v, %v; want %v", got, err, acc)
	}

	t.Run("the database keeps only the hash", func(t *testing.T) {
		var n int
		err := d.System.QueryRow(ctx, "SELECT count(*) FROM sessions WHERE token_hash = $1::text::bytea", sess.Token).Scan(&n)
		if err != nil || n != 0 {
			t.Fatalf("the plain token is in the table: %d, %v", n, err)
		}
	})

	for name, token := range map[string]string{"empty": "", "unknown": rand.Text()} {
		t.Run(name+" token", func(t *testing.T) {
			if _, err := s.SessionAccount(ctx, token); !errors.Is(err, ErrNoSession) {
				t.Fatalf("got %v, want ErrNoSession", err)
			}
		})
	}

	t.Run("an expired session does not work", func(t *testing.T) {
		old, err := s.StartSession(ctx, acc)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := d.System.Exec(ctx, "UPDATE sessions SET expires_at = now() - interval '1 second' WHERE token_hash = $1", tokenHash(old.Token)); err != nil {
			t.Fatal(err)
		}
		if _, err := s.SessionAccount(ctx, old.Token); !errors.Is(err, ErrNoSession) {
			t.Fatalf("got %v, want ErrNoSession", err)
		}
	})

	t.Run("logout ends one session", func(t *testing.T) {
		other, err := s.StartSession(ctx, acc)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.EndSession(ctx, acc, other.Token); err != nil {
			t.Fatal(err)
		}
		if _, err := s.SessionAccount(ctx, other.Token); !errors.Is(err, ErrNoSession) {
			t.Fatalf("ended session: %v", err)
		}
		if _, err := s.SessionAccount(ctx, sess.Token); err != nil {
			t.Fatalf("the other session must still work: %v", err)
		}
	})

	t.Run("end all sessions", func(t *testing.T) {
		if err := s.EndAllSessions(ctx, acc); err != nil {
			t.Fatal(err)
		}
		if _, err := s.SessionAccount(ctx, sess.Token); !errors.Is(err, ErrNoSession) {
			t.Fatalf("got %v, want ErrNoSession", err)
		}
	})
}

// The app role can call the lookup function, but cannot read sessions of
// other accounts directly.
func TestLookupFunctionIsTheOnlyWay(t *testing.T) {
	s, d := newService(t)
	ctx := t.Context()
	acc, _ := s.CreateAccount(ctx)
	if _, err := s.StartSession(ctx, acc); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := d.App.QueryRow(ctx, "SELECT count(*) FROM sessions").Scan(&n); err != nil || n != 0 {
		t.Fatalf("the app role sees %d sessions without an account context (%v)", n, err)
	}
}

func TestRequireSession(t *testing.T) {
	s, _ := newService(t)
	ctx := t.Context()
	acc, _ := s.CreateAccount(ctx)
	sess, err := s.StartSession(ctx, acc)
	if err != nil {
		t.Fatal(err)
	}

	var seen uuid.UUID
	h := s.RequireSession(slog.New(slog.DiscardHandler))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen, _ = AccountID(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))
	call := func(cookie string) int {
		req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/v1/dashboard/x", nil)
		if cookie != "" {
			req.AddCookie(&http.Cookie{Name: CookieName, Value: cookie})
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := call(sess.Token); code != http.StatusNoContent || seen != acc {
		t.Fatalf("live session: status %d, account %v", code, seen)
	}
	if code := call(""); code != http.StatusUnauthorized {
		t.Fatalf("no cookie: status %d", code)
	}
	if code := call("forged"); code != http.StatusUnauthorized {
		t.Fatalf("forged cookie: status %d", code)
	}
}

func TestCookie(t *testing.T) {
	s := New(nil, Options{SecureCookie: true})
	rec := httptest.NewRecorder()
	s.SetCookie(rec, Session{Token: "t", ExpiresAt: time.Now().Add(time.Hour)})
	c := rec.Result().Cookies()[0]
	if !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode || c.Path != "/" {
		t.Fatalf("cookie flags: %+v", c)
	}
}

// The lookup function goes around RLS, so its rights must be exact
// (ADR 0013): the system role owns it, PUBLIC cannot call it, the app can.
func TestLookupFunctionRights(t *testing.T) {
	_, d := newService(t)
	var owner string
	var definer, public, app bool
	err := d.Owner.QueryRow(t.Context(), `
		SELECT pg_get_userbyid(p.proowner), p.prosecdef,
		       EXISTS (SELECT 1 FROM aclexplode(p.proacl) a WHERE a.grantee = 0),
		       has_function_privilege('app_app', p.oid, 'EXECUTE')
		FROM pg_proc p WHERE p.proname = 'session_account'`).Scan(&owner, &definer, &public, &app)
	if err != nil {
		t.Fatal(err)
	}
	if owner != "app_system" || !definer || public || !app {
		t.Fatalf("owner %s, security definer %v, PUBLIC can execute %v, app can execute %v", owner, definer, public, app)
	}
}

// The scheduled clean-up deletes expired sessions of all accounts and keeps
// live ones. It uses the system role (ADR 0009).
func TestCleanupWorker(t *testing.T) {
	s, d := newService(t)
	ctx := t.Context()
	a, _ := s.CreateAccount(ctx)
	b, _ := s.CreateAccount(ctx)
	live, err := s.StartSession(ctx, a)
	if err != nil {
		t.Fatal(err)
	}
	expired, err := s.StartSession(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.System.Exec(ctx, "UPDATE sessions SET expires_at = now() - interval '1 second' WHERE token_hash = $1", tokenHash(expired.Token)); err != nil {
		t.Fatal(err)
	}

	if err := NewCleanupWorker(d.System, slog.New(slog.DiscardHandler)).Work(ctx, &river.Job[CleanupArgs]{}); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := d.System.QueryRow(ctx, "SELECT count(*) FROM sessions").Scan(&n); err != nil || n != 1 {
		t.Fatalf("sessions left = %d (%v), want 1", n, err)
	}
	if _, err := s.SessionAccount(ctx, live.Token); err != nil {
		t.Fatalf("the live session is gone: %v", err)
	}
}
