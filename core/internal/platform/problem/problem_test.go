package problem_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stilnar92/modern-project-template/core/internal/platform/openapi"
	"github.com/stilnar92/modern-project-template/core/internal/platform/problem"
)

func decode(t *testing.T, rec *httptest.ResponseRecorder) openapi.Problem {
	t.Helper()
	if got := rec.Header().Get("Content-Type"); got != problem.ContentType {
		t.Fatalf("Content-Type = %q, want %q", got, problem.ContentType)
	}
	var p openapi.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("body is not a problem: %v: %s", err, rec.Body)
	}
	return p
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name   string
		status int
		detail string
	}{
		{"with detail", http.StatusBadRequest, "The request is not valid."},
		{"without detail", http.StatusInternalServerError, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			problem.Write(rec, tt.status, tt.detail)

			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d", rec.Code, tt.status)
			}
			p := decode(t, rec)
			if p.Type != openapi.AboutBlank || p.Status != tt.status || p.Title != http.StatusText(tt.status) {
				t.Fatalf("problem = %+v", p)
			}
			if (p.Detail != nil) != (tt.detail != "") {
				t.Fatalf("detail = %v, want %q", p.Detail, tt.detail)
			}
		})
	}
}

func TestRoutes(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := problem.Routes(mux)

	t.Run("a known route passes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil))
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})

	t.Run("an unknown path is a 404 problem", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/nothing", nil))
		if p := decode(t, rec); rec.Code != http.StatusNotFound || p.Status != http.StatusNotFound {
			t.Fatalf("status = %d, problem = %+v", rec.Code, p)
		}
	})

	t.Run("another method is a 405 problem with Allow", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/healthz", nil))
		if p := decode(t, rec); rec.Code != http.StatusMethodNotAllowed || p.Status != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, problem = %+v", rec.Code, p)
		}
		if allow := rec.Header().Get("Allow"); allow == "" {
			t.Fatal("no Allow header")
		}
	})
}
