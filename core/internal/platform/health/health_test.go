package health

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stilnar92/modern-project-template/core/internal/platform/openapi"
	"github.com/stilnar92/modern-project-template/core/internal/platform/testdb"
)

func get(t *testing.T, s Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	handler := openapi.HandlerFromMux(openapi.NewStrictHandler(s, nil), http.NewServeMux())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
	return rec
}

func TestGetHealth(t *testing.T) {
	rec := get(t, Server{}, "/healthz")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got, want := rec.Body.String(), "{\"status\":\"ok\"}\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestGetReady(t *testing.T) {
	db := testdb.Start(t)
	quiet := slog.New(slog.DiscardHandler)

	t.Run("the database answers", func(t *testing.T) {
		rec := get(t, Server{DB: db.App, Logger: quiet}, "/readyz")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
		}
	})

	t.Run("the database does not answer", func(t *testing.T) {
		closed, err := pgxpool.New(t.Context(), db.App.Config().ConnString())
		if err != nil {
			t.Fatal(err)
		}
		closed.Close()

		rec := get(t, Server{DB: closed, Logger: quiet}, "/readyz")
		if rec.Code != http.StatusServiceUnavailable || rec.Header().Get("Content-Type") != "application/problem+json" {
			t.Fatalf("status = %d, Content-Type %q", rec.Code, rec.Header().Get("Content-Type"))
		}
		if strings.Contains(rec.Body.String(), "closed") {
			t.Fatalf("the response shows the internal error: %s", rec.Body)
		}
	})
}

func TestCheck(t *testing.T) {
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }))
	defer srv.Close()

	if err := Check(t.Context(), srv.URL); err != nil {
		t.Fatalf("200: got %v, want nil", err)
	}
	status = http.StatusInternalServerError
	if err := Check(t.Context(), srv.URL); err == nil {
		t.Fatal("500: got nil, want an error")
	}
	srv.Close()
	if err := Check(t.Context(), srv.URL); err == nil {
		t.Fatal("no server: got nil, want an error")
	}
}
