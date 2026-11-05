package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stilnar92/modern-project-template/core/internal/platform/health"
	"github.com/stilnar92/modern-project-template/core/internal/platform/openapi"
	"github.com/stilnar92/modern-project-template/core/internal/platform/problem"
)

// failing returns an error with internal details, as a broken query would.
type failing struct{}

func (failing) GetHealth(context.Context, openapi.GetHealthRequestObject) (openapi.GetHealthResponseObject, error) {
	return nil, errors.New(`pq: relation "accounts" does not exist`)
}

func TestCheckHealth(t *testing.T) {
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }))
	defer srv.Close()

	if err := checkHealth(t.Context(), srv.URL); err != nil {
		t.Fatalf("200: got %v, want nil", err)
	}
	status = http.StatusInternalServerError
	if err := checkHealth(t.Context(), srv.URL); err == nil {
		t.Fatal("500: got nil, want an error")
	}
	srv.Close()
	if err := checkHealth(t.Context(), srv.URL); err == nil {
		t.Fatal("no server: got nil, want an error")
	}
}

func serve(t *testing.T, handler http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), method, path, nil))
	return rec
}

func TestHandler(t *testing.T) {
	quiet := slog.New(slog.DiscardHandler)

	t.Run("the health check answers", func(t *testing.T) {
		rec := serve(t, newHandler(quiet, health.Server{}), http.MethodGet, "/healthz")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("a handler error is a 500 problem without internal details", func(t *testing.T) {
		var log bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&log, nil))
		rec := serve(t, newHandler(logger, failing{}), http.MethodGet, "/healthz")

		if rec.Code != http.StatusInternalServerError || rec.Header().Get("Content-Type") != problem.ContentType {
			t.Fatalf("status = %d, Content-Type = %q", rec.Code, rec.Header().Get("Content-Type"))
		}
		if strings.Contains(rec.Body.String(), "accounts") {
			t.Fatalf("the response shows the internal error: %s", rec.Body)
		}
		if !strings.Contains(log.String(), "accounts") {
			t.Fatalf("the log does not have the internal error: %s", log.String())
		}
	})

	for _, tt := range []struct {
		method, path string
		status       int
	}{
		{http.MethodGet, "/nothing", http.StatusNotFound},
		{http.MethodPost, "/healthz", http.StatusMethodNotAllowed},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := serve(t, newHandler(quiet, health.Server{}), tt.method, tt.path)
			if rec.Code != tt.status || rec.Header().Get("Content-Type") != problem.ContentType {
				t.Fatalf("status = %d, Content-Type = %q", rec.Code, rec.Header().Get("Content-Type"))
			}
		})
	}
}
