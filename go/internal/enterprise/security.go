package enterprise

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// SecurityProvider defines the interface for enterprise-grade security features.
type SecurityProvider interface {
	ValidateSSO(ctx context.Context, token string) (bool, error)
	Authorize(ctx context.Context, userID string, resource string, action string) (bool, error)
}

// EnterpriseWrapper wraps the core execution engine with enterprise security.
type EnterpriseWrapper struct {
	provider   SecurityProvider
	configPath string
	mu         sync.RWMutex
	ssoConfig  map[string]string
	roles      []map[string]any
}

type EnterpriseConfig struct {
	SSO   map[string]string `json:"sso"`
	Roles []map[string]any  `json:"roles"`
}

// NewEnterpriseWrapper creates a new wrapper with the given provider.
func NewEnterpriseWrapper(provider SecurityProvider, workspaceRoot string) *EnterpriseWrapper {
	cfgPath := filepath.Join(workspaceRoot, ".tormentnexus", "enterprise_config.json")
	ew := &EnterpriseWrapper{
		provider:   provider,
		configPath: cfgPath,
		ssoConfig:  make(map[string]string),
		roles:      defaultRoles(),
	}
	_ = ew.Load()
	return ew
}

func defaultRoles() []map[string]any {
	return []map[string]any{
		{"name": "admin", "description": "Full system access", "permissions": []string{"read", "write", "admin", "audit"}},
		{"name": "operator", "description": "Daily operations", "permissions": []string{"read", "write", "execute"}},
		{"name": "viewer", "description": "Read-only access", "permissions": []string{"read"}},
	}
}

// Load loads the configuration from disk.
func (ew *EnterpriseWrapper) Load() error {
	ew.mu.Lock()
	defer ew.mu.Unlock()

	data, err := os.ReadFile(ew.configPath)
	if err != nil {
		return err
	}

	var config EnterpriseConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	if config.SSO != nil {
		ew.ssoConfig = config.SSO
	}
	if config.Roles != nil {
		ew.roles = config.Roles
	}
	return nil
}

// Save saves the current configuration to disk.
func (ew *EnterpriseWrapper) Save() error {
	ew.mu.RLock()
	config := EnterpriseConfig{
		SSO:   ew.ssoConfig,
		Roles: ew.roles,
	}
	ew.mu.RUnlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(ew.configPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(ew.configPath, data, 0644)
}

// Info returns enterprise license and security info.
func (ew *EnterpriseWrapper) Info() map[string]any {
	ew.mu.RLock()
	defer ew.mu.RUnlock()

	return map[string]any{
		"valid":       true,
		"licensedTo":  "TormentNexus Enterprise",
		"tier":        "enterprise",
		"maxNodes":    10,
		"features":    []string{"sso", "rbac", "audit", "encryption"},
		"expiresAt":   "",
		"ssoSettings": ew.ssoConfig,
	}
}

// GetRoles returns the available RBAC roles.
func (ew *EnterpriseWrapper) GetRoles() []map[string]any {
	ew.mu.RLock()
	defer ew.mu.RUnlock()
	return ew.roles
}

// UpdateSSO updates the SSO configuration.
func (ew *EnterpriseWrapper) UpdateSSO(sso map[string]string) error {
	ew.mu.Lock()
	ew.ssoConfig = sso
	ew.mu.Unlock()
	return ew.Save()
}

// UpdateRoles updates the RBAC roles.
func (ew *EnterpriseWrapper) UpdateRoles(roles []map[string]any) error {
	ew.mu.Lock()
	ew.roles = roles
	ew.mu.Unlock()
	return ew.Save()
}

// Middleware provides an HTTP middleware for enterprise security checks.
func (ew *EnterpriseWrapper) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Enterprise-SSO")
		if token != "" && ew.provider != nil {
			valid, err := ew.provider.ValidateSSO(r.Context(), token)
			if err != nil || !valid {
				http.Error(w, "Unauthorized: Invalid SSO token", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
