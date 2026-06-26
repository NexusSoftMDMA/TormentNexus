// Package logging provides structured logging using slog with
// correlation IDs, component tags, and request-scoped context propagation.
package logging

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

// contextKey is a private type for context value keys.
type contextKey string

const (
	// CtxRequestID is the context key for the request correlation ID.
	CtxRequestID contextKey = "request_id"
	// CtxComponent is the context key for the component name.
	CtxComponent contextKey = "component"
)

var (
	global   *slog.Logger
	initOnce sync.Once
)

// Init sets up the global structured logger.
// format: "json" or "text". Default is JSON.
// level: "debug", "info", "warn", "error". Default is "info".
func Init(format, level string) {
	initOnce.Do(func() {
		var lvl slog.Level
		switch level {
		case "debug":
			lvl = slog.LevelDebug
		case "warn":
			lvl = slog.LevelWarn
		case "error":
			lvl = slog.LevelError
		default:
			lvl = slog.LevelInfo
		}

		var handler slog.Handler
		opts := &slog.HandlerOptions{Level: lvl}
		switch format {
		case "text":
			handler = slog.NewTextHandler(os.Stderr, opts)
		default:
			handler = slog.NewJSONHandler(os.Stderr, opts)
		}
		global = slog.New(handler)

		// Set as default slog logger too
		slog.SetDefault(global)
	})
}

// Logger returns the global logger, initializing with defaults if not set.
func Logger() *slog.Logger {
	initOnce.Do(func() {
		Init("json", "info")
	})
	return global
}

// WithRequestID generates a random request ID and returns a new context and logger.
func WithRequestID(ctx context.Context) (context.Context, *slog.Logger) {
	id := generateID()
	ctx = context.WithValue(ctx, CtxRequestID, id)
	return ctx, FromContext(ctx).With("request_id", id)
}

// FromContext returns a logger enriched with context values (request_id, component).
func FromContext(ctx context.Context) *slog.Logger {
	logger := Logger()
	if ctx == nil {
		return logger
	}
	var attrs []slog.Attr
	if id, ok := ctx.Value(CtxRequestID).(string); ok && id != "" {
		attrs = append(attrs, slog.String("request_id", id))
	}
	if comp, ok := ctx.Value(CtxComponent).(string); ok && comp != "" {
		attrs = append(attrs, slog.String("component", comp))
	}
	if len(attrs) > 0 {
		return slog.New(logger.Handler().WithAttrs(attrs))
	}
	return logger
}

// WithComponent returns a context and logger scoped to a component.
func WithComponent(ctx context.Context, component string) (context.Context, *slog.Logger) {
	ctx = context.WithValue(ctx, CtxComponent, component)
	return ctx, FromContext(ctx)
}

// generateID creates a short random hex string for request correlation.
func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
