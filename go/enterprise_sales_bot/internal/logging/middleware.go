package logging

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Middleware wraps an http.Handler with request logging and correlation IDs.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Pick up request ID from header or generate one
		rid := r.Header.Get("X-Request-ID")
		ctx := r.Context()
		if rid != "" {
			ctx = context.WithValue(ctx, CtxRequestID, rid)
		} else {
			var logCtx context.Context
			ctx, _ = WithRequestID(ctx)
			_ = logCtx // suppress unused warning
		}

		// Extract existing or generate request ID
		ctx, _ = WithRequestID(ctx)
		r = r.WithContext(ctx)

		// Wrap response writer to capture status
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Log request start
		logger := FromContext(ctx)
		logger.DebugContext(ctx, "request started",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
		)

		// Serve
		next.ServeHTTP(rw, r)

		// Log request end
		duration := time.Since(start)
		level := slog.LevelInfo
		if rw.statusCode >= 500 {
			level = slog.LevelError
		} else if rw.statusCode >= 400 {
			level = slog.LevelWarn
		}
		logger.LogAttrs(ctx, level, "request completed",
			slog.Int("status", rw.statusCode),
			slog.Duration("duration", duration),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)
	})
}

// RequestIDFromContext extracts the request ID from a context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(CtxRequestID).(string); ok {
		return id
	}
	return ""
}
