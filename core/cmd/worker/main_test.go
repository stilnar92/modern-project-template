package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stilnar92/modern-project-template/core/internal/platform/testdb"
)

// Each job kind is registered, and River accepts the configuration.
func TestNewRiver(t *testing.T) {
	d := testdb.Start(t)
	if _, err := newRiver(d.App, d.System, slog.New(slog.DiscardHandler), 2); err != nil {
		t.Fatal(err)
	}
}

// /healthz answers 503 before the River client starts, 200 while it runs,
// and 503 after it stops.
func TestLiveness(t *testing.T) {
	status := func(l *liveness) int {
		rec := httptest.NewRecorder()
		l.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil))
		return rec.Code
	}
	l := &liveness{}
	if got := status(l); got != http.StatusServiceUnavailable {
		t.Fatalf("before start: status = %d, want 503", got)
	}
	stopped := make(chan struct{})
	l.running(stopped)
	if got := status(l); got != http.StatusOK {
		t.Fatalf("running: status = %d, want 200", got)
	}
	close(stopped)
	if got := status(l); got != http.StatusServiceUnavailable {
		t.Fatalf("stopped: status = %d, want 503", got)
	}
}
