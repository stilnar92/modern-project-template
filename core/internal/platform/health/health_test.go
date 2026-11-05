package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stilnar92/modern-project-template/core/internal/platform/openapi"
)

func TestGetHealth(t *testing.T) {
	handler := openapi.HandlerFromMux(openapi.NewStrictHandler(Server{}, nil), http.NewServeMux())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got, want := rec.Body.String(), "{\"status\":\"ok\"}\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}
