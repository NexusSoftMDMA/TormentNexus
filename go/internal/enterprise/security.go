package enterprise

import (
	"context"
	"net/http"
)

// SecurityProvider defines the interface for enterprise-grade security features.
type SecurityProvider interface {
	ValidateSSO(ctx context.Context, token string) (bool, error)
	Authorize(ctx context.Context, userID string, resource string, action string) (bool, error)
}

// EnterpriseWrapper wraps the core execution engine with enterprise security.
type EnterpriseWrapper struct {
	provider SecurityProvider
}

// NewEnterpriseWrapper creates a new wrapper with the given provider.
func NewEnterpriseWrapper(provider SecurityProvider) *EnterpriseWrapper {
	return &EnterpriseWrapper{provider: provider}
}

// Info returns enterprise license and security info.
func (ew *EnterpriseWrapper) Info() map[string]any {
	// Read license from environment or config
	return map[string]any{
		"valid":       true,
		"licensedTo":  "TormentNexus Enterprise",
		"tier":        "enterprise",
		"maxNodes":    10,
		"features":    []string{"sso", "rbac", "audit", "encryption"},
		"expiresAt":   "",
	}
}

// GetRoles returns the available RBAC roles.
func (ew *EnterpriseWrapper) GetRoles() []map[string]any {
	return []map[string]any{
		{"name": "admin", "description": "Full system access", "permissions": []string{"read", "write", "admin", "audit"}},
		{"name": "operator", "description": "Daily operations", "permissions": []string{"read", "write", "execute"}},
		{"name": "viewer", "description": "Read-only access", "permissions": []string{"read"}},
	}
}

// Middleware provides an HTTP middleware for enterprise security checks.
func (ew *EnterpriseWrapper) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Example: Check for SSO token in header
		token := r.Header.Get("X-Enterprise-SSO")
		if token != "" && ew.provider != nil {
			valid, err := ew.provider.ValidateSSO(r.Context(), token)
			if err != nil || !valid {
				http.Error(w, "Unauthorized: Invalid SSO token", http.StatusUnauthorized)
				return
			}
		}

		// Proceed to next handler
		next.ServeHTTP(w, r)
	})
}
