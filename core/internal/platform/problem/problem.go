// Package problem writes error responses in the Problem Details format
// (RFC 9457, backend standard 02). A response never contains internal
// details: the real error goes only to the log and to error tracking.
package problem

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/stilnar92/modern-project-template/core/internal/platform/observability"
	"github.com/stilnar92/modern-project-template/core/internal/platform/openapi"
)

// ContentType is the media type of a problem response.
const ContentType = "application/problem+json"

// Write sends a plain HTTP error: the type "about:blank" and the name of the
// status as the title. The detail must be safe to show to the client.
func Write(w http.ResponseWriter, status int, detail string) {
	p := openapi.Problem{Type: openapi.AboutBlank, Title: http.StatusText(status), Status: status}
	if detail != "" {
		p.Detail = &detail
	}
	w.Header().Set("Content-Type", ContentType)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(p)
}

// RequestError answers a request that the server cannot read, for example a
// wrong parameter or a body that is not JSON. The error text goes to the log.
func RequestError(logger *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		logger.InfoContext(r.Context(), "invalid request", "method", r.Method, "path", r.URL.Path, "err", err)
		Write(w, http.StatusBadRequest, "The request is not valid.")
	}
}

// ServerError answers an error that a handler returns. The error goes to the
// log and to error tracking, and the client gets only the status.
func ServerError(logger *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		observability.ReportError(logger.With("method", r.Method, "path", r.URL.Path), "request failed", err)
		Write(w, http.StatusInternalServerError, "")
	}
}

// Routes wraps the router so that a path without a route (404) and a known
// path with another method (405) get a problem, not the plain text of
// http.ServeMux.
func Routes(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pattern := mux.Handler(r); pattern != "" {
			mux.ServeHTTP(w, r)
			return
		}
		// The mux answers 404, or 405 with an Allow header. Keep the status
		// and the header, and replace the body.
		rec := &statusRecorder{header: http.Header{}}
		mux.ServeHTTP(rec, r)
		if allow := rec.header.Get("Allow"); allow != "" {
			w.Header().Set("Allow", allow)
		}
		Write(w, rec.status, "")
	})
}

// statusRecorder keeps the status and the headers and drops the body.
type statusRecorder struct {
	header http.Header
	status int
}

func (s *statusRecorder) Header() http.Header         { return s.header }
func (s *statusRecorder) WriteHeader(status int)      { s.status = status }
func (s *statusRecorder) Write(b []byte) (int, error) { return len(b), nil }
