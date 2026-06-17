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
