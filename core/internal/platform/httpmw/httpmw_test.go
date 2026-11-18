package httpmw

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stilnar92/modern-project-template/core/internal/platform/observability"
	"github.com/stilnar92/modern-project-template/core/internal/platform/problem"
)

func serve(t *testing.T, h http.Handler, header string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x?phone=79990001122", nil)
	if header != "" {
		req.Header.Set(RequestIDHeader, header)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// echoID writes the request ID of the context into the body.
var echoID = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(observability.RequestID(r.Context())))
})

func TestRequestID(t *testing.T) {
	t.Run("a valid client ID is kept", func(t *testing.T) {
		rec := serve(t, RequestID(echoID), "abc-123")
		if rec.Body.String() != "abc-123" || rec.Header().Get(RequestIDHeader) != "abc-123" {
			t.Fatalf("body %q, header %q", rec.Body, rec.Header().Get(RequestIDHeader))
		}
	})
	for name, header := range map[string]string{
		"no ID":         "",
		"a log forgery": "x\n{\"level\":\"ERROR\"}",
		"too long":      strings.Repeat("a", 65),
	} {
		t.Run(name+" gets a new ID", func(t *testing.T) {
			rec := serve(t, RequestID(echoID), header)
			id := rec.Body.String()
			if id == "" || id == header || rec.Header().Get(RequestIDHeader) != id {
				t.Fatalf("body %q, header %q", id, rec.Header().Get(RequestIDHeader))
			}
		})
	}
}

func TestAccessLog(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(observability.NewContextHandler(slog.NewJSONHandler(&buf, nil)))
	teapot := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })

	serve(t, RequestID(AccessLog(logger)(teapot)), "req-1")

	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	if rec["status"] != float64(http.StatusTeapot) || rec["path"] != "/x" || rec["request_id"] != "req-1" {
		t.Fatalf("record: %v", rec)
	}
	if strings.Contains(buf.String(), "7999") {
		t.Fatalf("the query string is in the log: %s", buf.String())
	}
}

func TestAccessLogSystemChecks(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	down := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) })

	for name, tt := range map[string]struct {
		h      http.Handler
		path   string
		logged bool
	}{
		"a passing check is not logged at Info": {ok, "/healthz", false},
		"a failing check is logged":             {down, "/readyz", true},
		"other paths are logged":                {ok, "/v1/x", true},
	} {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil)) // Info level
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil)
			AccessLog(logger)(tt.h).ServeHTTP(httptest.NewRecorder(), req)
			if got := buf.Len() > 0; got != tt.logged {
				t.Fatalf("logged = %v, want %v: %s", got, tt.logged, buf.String())
			}
		})
	}
}

func TestRecover(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	boom := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("secret internal state") })

	rec := serve(t, Recover(logger)(boom), "")

	if rec.Code != http.StatusInternalServerError || rec.Header().Get("Content-Type") != problem.ContentType {
		t.Fatalf("status %d, Content-Type %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("the panic is in the response: %s", rec.Body)
	}
	if !strings.Contains(buf.String(), "secret internal state") || !strings.Contains(buf.String(), "goroutine") {
		t.Fatalf("the log has no panic and stack: %s", buf.String())
	}
}

func TestRecoverPassesAbortHandler(t *testing.T) {
	abort := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) })
	defer func() {
		if v := recover(); v != http.ErrAbortHandler { //nolint:errorlint // the same comparison as the server
			t.Fatalf("recovered %v, want http.ErrAbortHandler", v)
		}
	}()
	serve(t, Recover(slog.New(slog.DiscardHandler))(abort), "")
}
