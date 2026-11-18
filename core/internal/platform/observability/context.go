package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type requestIDKey struct{}

// WithRequestID returns a context that carries the request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestID returns the request ID of the context, or "".
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

type jobKey struct{}

type jobInfo struct {
	id   int64
	kind string
}

// WithJob returns a context that carries the background job (backend
// standard 06: "the job ID for background work").
func WithJob(ctx context.Context, id int64, kind string) context.Context {
	return context.WithValue(ctx, jobKey{}, jobInfo{id: id, kind: kind})
}

// ContextHandler adds request_id, job_id, job_kind and trace_id from the
// context to each record (backend standard 06). Use the *Context log methods, or the IDs
// are lost.
type ContextHandler struct{ slog.Handler }

// NewContextHandler wraps h.
func NewContextHandler(h slog.Handler) *ContextHandler { return &ContextHandler{h} }

// Handle adds the IDs and passes the record on.
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := RequestID(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	if j, ok := ctx.Value(jobKey{}).(jobInfo); ok {
		r.AddAttrs(slog.Int64("job_id", j.id), slog.String("job_kind", j.kind))
	}
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(slog.String("trace_id", sc.TraceID().String()), slog.String("span_id", sc.SpanID().String()))
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs keeps the wrapper.
func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{h.Handler.WithAttrs(attrs)}
}

// WithGroup keeps the wrapper.
func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{h.Handler.WithGroup(name)}
}
