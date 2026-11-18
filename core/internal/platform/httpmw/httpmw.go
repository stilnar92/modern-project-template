// Package httpmw holds the HTTP middleware of the core (ADR 0012): the
// request ID, the access log and panic recovery. The api applies them in
// this order, from the outside in:
//
//	otelhttp → RequestID → AccessLog → Recover → routes
//
// So the access log sees the 500 of a recovered panic, and each log record
// has the request ID and the trace ID.
package httpmw

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/felixge/httpsnoop"

	"github.com/stilnar92/modern-project-template/core/internal/platform/observability"
	"github.com/stilnar92/modern-project-template/core/internal/platform/problem"
)

// RequestIDHeader carries the request ID in both directions.
const RequestIDHeader = "X-Request-ID"

// A request ID from the client is accepted only in this form, so a client
// cannot put arbitrary text into the logs.
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// RequestID keeps a valid X-Request-ID of the request or makes a new one. It
// puts the ID into the context and into the response header, so a user can
// give it to support and we can find the log records.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if !validRequestID.MatchString(id) {
			id = rand.Text()
		}
		w.Header().Set(RequestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(observability.WithRequestID(r.Context(), id)))
	})
}

// systemChecks are called every few seconds by Docker and the proxy. Their
// records go to the Debug level, so they do not fill the log.
var systemChecks = map[string]bool{"/healthz": true, "/readyz": true}

// AccessLog writes one record for each request. It does not log the query
// string or the body: they can contain personal data (backend standard 06).
func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m := httpsnoop.CaptureMetrics(next, w, r)
			level := slog.LevelInfo
			if systemChecks[r.URL.Path] && m.Code < http.StatusInternalServerError {
				level = slog.LevelDebug
			}
			logger.Log(r.Context(), level, "request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", m.Code,
				"bytes", m.Written,
				"duration_ms", float64(m.Duration)/float64(time.Millisecond),
			)
		})
	}
}

// Recover turns a panic in a handler into a 500 Problem response. The panic
// and its stack go to the log and to Sentry, never to the client.
// http.ErrAbortHandler passes through: the server uses it to stop a response.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				if v == http.ErrAbortHandler { //nolint:errorlint // the server compares the value, not a wrapped error
					panic(v)
				}
				err := fmt.Errorf("panic: %v\n%s", v, debug.Stack())
				observability.ReportError(r.Context(), logger.With("method", r.Method, "path", r.URL.Path), "handler panic", err)
				problem.Write(w, http.StatusInternalServerError, "")
			}()
			next.ServeHTTP(w, r)
		})
	}
}
