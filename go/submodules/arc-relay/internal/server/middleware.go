package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/comma-compliance/arc-relay/internal/store"
)

type contextKey string

const userContextKey contextKey = "user"

// UserFromContext retrieves the authenticated user from the request context.
func UserFromContext(ctx context.Context) *store.User {
	u, _ := ctx.Value(userContextKey).(*store.User)
	return u
}

// setWWWAuthenticate adds the RFC 9728 WWW-Authenticate header for OAuth discovery.
// The resource_metadata URL includes the request path so the metadata's resource
// field matches the exact URL the client probed (required by RFC 9728).
func setWWWAuthenticate(w http.ResponseWriter, baseURL string, r *http.Request) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(
		`Bearer resource_metadata="%s/.well-known/oauth-protected-resource%s"`, baseURL, r.URL.Path))
}

// jsonError writes a JSON error response with proper Content-Type.
// Unlike http.Error(), this ensures MCP/OAuth clients see application/json.
func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = fmt.Fprint(w, msg)
}

// APIKeyAuth middleware validates Bearer token API keys.
// Used for management API routes (/api/servers) - does NOT accept OAuth tokens.
func APIKeyAuth(users *store.UserStore, baseURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				setWWWAuthenticate(w, baseURL, r)
				jsonError(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(auth, "Bearer ") {
				setWWWAuthenticate(w, baseURL, r)
				jsonError(w, `{"error":"invalid Authorization header, expected Bearer token"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(auth, "Bearer ")
			user, err := users.ValidateAPIKey(token)
			if err != nil {
				slog.Warn("auth: validate api key failed", "path", r.URL.Path, "remote", r.RemoteAddr, "err", err)
				jsonError(w, `{"error":"internal error"}`, http.StatusInternalServerError)
				return
			}
			if user == nil {
				setWWWAuthenticate(w, baseURL, r)
				jsonError(w, `{"error":"invalid or revoked API key"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MCPAuth middleware validates Bearer tokens for MCP proxy routes.
// Checks API keys first, then OAuth tokens. This ensures OAuth tokens
// only work on /mcp/ proxy routes (not management API).
func MCPAuth(users *store.UserStore, oauthTokens *store.OAuthTokenStore, baseURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				setWWWAuthenticate(w, baseURL, r)
				jsonError(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(auth, "Bearer ") {
				setWWWAuthenticate(w, baseURL, r)
				jsonError(w, `{"error":"invalid Authorization header, expected Bearer token"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(auth, "Bearer ")

			// Try API key first
			user, err := users.ValidateAPIKey(token)
			if err != nil {
				slog.Warn("auth: validate api key failed", "path", r.URL.Path, "remote", r.RemoteAddr, "err", err)
				jsonError(w, `{"error":"internal error"}`, http.StatusInternalServerError)
				return
			}

			// If not an API key, try OAuth token
			if user == nil {
				user, err = oauthTokens.Validate(token)
				if err != nil {
					slog.Warn("auth: validate oauth token failed", "path", r.URL.Path, "remote", r.RemoteAddr, "err", err)
					jsonError(w, `{"error":"internal error"}`, http.StatusInternalServerError)
					return
				}
			}

			if user == nil {
				setWWWAuthenticate(w, baseURL, r)
				jsonError(w, `{"error":"invalid or revoked token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminOnly middleware ensures the user has admin role.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil || user.Role != "admin" {
			jsonError(w, `{"error":"admin access required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireWriteAccess checks that the authenticated user has write or admin access level.
// Returns true if access is granted, false if a 403 was sent.
func requireWriteAccess(w http.ResponseWriter, r *http.Request) bool {
	user := UserFromContext(r.Context())
	if user == nil {
		jsonError(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
		return false
	}
	if user.AccessLevel != "write" && user.AccessLevel != "admin" {
		jsonError(w, `{"error":"write access required"}`, http.StatusForbidden)
		return false
	}
	return true
}

// requireAdminAccess checks that the authenticated user has admin access level.
// Returns true if access is granted, false if a 403 was sent.
func requireAdminAccess(w http.ResponseWriter, r *http.Request) bool {
	user := UserFromContext(r.Context())
	if user == nil {
		jsonError(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
		return false
	}
	if user.AccessLevel != "admin" {
		jsonError(w, `{"error":"admin access required"}`, http.StatusForbidden)
		return false
	}
	return true
}
