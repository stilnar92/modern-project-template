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
	"github.com/stilnar92/modern-project-template/core/internal/platform/httpmw"
	"github.com/stilnar92/modern-project-template/core/internal/platform/openapi"
	"github.com/stilnar92/modern-project-template/core/internal/platform/problem"
)

// failing returns an error with internal details, as a broken query would.
type failing struct{}

func (failing) GetHealth(context.Context, openapi.GetHealthRequestObject) (openapi.GetHealthResponseObject, error) {
	return nil, errors.New(`pq: relation "accounts" does not exist`)
}

func (failing) GetReady(context.Context, openapi.GetReadyRequestObject) (openapi.GetReadyResponseObject, error) {
	return nil, errors.New("not used")
}

// panicking panics in a handler, as a nil pointer would.
type panicking struct{ failing }

func (panicking) GetHealth(context.Context, openapi.GetHealthRequestObject) (openapi.GetHealthResponseObject, error) {
	panic("secret internal state")
}

func serve(t *testing.T, handler http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), method, path, nil))
	return rec
}

func TestHandler(t *testing.T) {
	quiet := slog.New(slog.DiscardHandler)

	t.Run("the health check answers with a request ID", func(t *testing.T) {
		rec := serve(t, newHandler(quiet, health.Server{}), http.MethodGet, "/healthz")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if rec.Header().Get(httpmw.RequestIDHeader) == "" {
			t.Fatal("the response has no request ID")
		}
	})

	t.Run("a panic is a 500 problem, and the access log has it", func(t *testing.T) {
		var log bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&log, nil))
		rec := serve(t, newHandler(logger, panicking{}), http.MethodGet, "/healthz")

		if rec.Code != http.StatusInternalServerError || rec.Header().Get("Content-Type") != problem.ContentType {
			t.Fatalf("status = %d, Content-Type = %q", rec.Code, rec.Header().Get("Content-Type"))
		}
		if strings.Contains(rec.Body.String(), "secret") {
			t.Fatalf("the response shows the panic: %s", rec.Body)
		}
		if !strings.Contains(log.String(), `"status":500`) {
			t.Fatalf("the access log has no 500: %s", log.String())
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

	t.Run("a cross-site POST is refused with a problem", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/healthz", nil)
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		rec := httptest.NewRecorder()
		newHandler(quiet, health.Server{}).ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden || rec.Header().Get("Content-Type") != problem.ContentType {
			t.Fatalf("status = %d, Content-Type = %q", rec.Code, rec.Header().Get("Content-Type"))
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
