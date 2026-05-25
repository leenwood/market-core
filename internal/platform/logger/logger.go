package logger

import (
	"context"
	"log/slog"
	"os"
)

type requestIDKey struct{}
type traceIDKey struct{}

func New(level, format string) *slog.Logger {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: l}
	var h slog.Handler
	if format == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(h)
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, id)
}

func TraceIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey{}).(string); ok {
		return id
	}
	return ""
}

// FromContext returns base logger enriched with request_id and trace_id from ctx.
func FromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	l := base
	if rid := RequestIDFromContext(ctx); rid != "" {
		l = l.With("request_id", rid)
	}
	if tid := TraceIDFromContext(ctx); tid != "" {
		l = l.With("trace_id", tid)
	}
	return l
}
