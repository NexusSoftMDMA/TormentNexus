package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ErrSlugConflict is returned when a slug name is already taken by another server.
var ErrSlugConflict = errors.New("server slug already exists")

// slugPattern matches valid slug names: lowercase alphanumeric with hyphens, min 2 chars.
var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

// ValidateSlug checks that a slug name meets the required format.
func ValidateSlug(name string) error {
	if name == "" {
		return fmt.Errorf("slug name is required")
	}
	if !slugPattern.MatchString(name) {
		return fmt.Errorf("slug must be lowercase alphanumeric with hyphens, at least 2 characters (got %q)", name)
	}
	return nil
}

type ServerType string

const (
	ServerTypeStdio  ServerType = "stdio"
	ServerTypeHTTP   ServerType = "http"
	ServerTypeRemote ServerType = "remote"
)

type ServerStatus string

const (
	StatusStopped  ServerStatus = "stopped"
	StatusStarting ServerStatus = "starting"
	StatusRunning  ServerStatus = "running"
	StatusError    ServerStatus = "error"
)

type HealthStatus string

const (
	HealthHealthy   HealthStatus = "healthy"
	HealthUnhealthy HealthStatus = "unhealthy"
	HealthUnknown   HealthStatus = "unknown"
)

type Server struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	DisplayName     string          `json:"display_name"`
	ServerType      ServerType      `json:"server_type"`
	Config          json.RawMessage `json:"config"`
	Status          ServerStatus    `json:"status"`
	ErrorMsg        string          `json:"error_msg,omitempty"`
	Health          HealthStatus    `json:"health"`
	HealthCheckAt   *time.Time      `json:"health_check_at,omitempty"`
	HealthError     string          `json:"health_error,omitempty"`
	OptimizeEnabled bool            `json:"optimize_enabled"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// StdioConfig holds config for Docker-managed stdio servers.
type StdioConfig struct {
	Image      string            `json:"image"`
	Build      *StdioBuildConfig `json:"build,omitempty"`
	Entrypoint []string          `json:"entrypoint,omitempty"`
	Command    []string          `json:"command,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
}

// StdioBuildConfig holds metadata for auto-building a Docker image from a package.
type StdioBuildConfig struct {
	Runtime    string `json:"runtime"`              // "python" or "node"
	Package    string `json:"package"`              // pip/npm package name
	Version    string `json:"version,omitempty"`    // package version (empty = latest)
	GitURL     string `json:"git_url,omitempty"`    // alternative: build from git repo
	GitRef     string `json:"git_ref,omitempty"`    // branch, tag, or commit hash
	Dockerfile string `json:"dockerfile,omitempty"` // alternative: custom Dockerfile text
}

// BuildImageTag returns the Docker image tag for an auto-built image.
func (b *StdioBuildConfig) BuildImageTag() string {
	source := b.Package
	if source == "" && b.GitURL != "" {
		source = repoNameFromURL(b.GitURL)
	}
	name := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.' || r == '/' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r + 32 // lowercase
		}
		return '-'
	}, source)
	name = strings.Trim(name, "-")
	version := b.Version
	if version == "" && b.GitRef != "" {
		// Sanitize git ref for Docker tag safety (replace / with -)
		version = strings.ReplaceAll(b.GitRef, "/", "-")
	}
	if version == "" {
		version = "latest"
	}
	return fmt.Sprintf("arc-relay-build/%s:%s", name, version)
}

// repoNameFromURL extracts the repository name from a git URL.
// Handles both https://github.com/user/repo.git and git@host:user/repo.git formats.
func repoNameFromURL(u string) string {
	// Handle scp-style: git@host:user/repo.git
	if idx := strings.LastIndex(u, ":"); idx > 0 && !strings.Contains(u, "://") {
		u = u[idx+1:]
	}
	// Strip trailing .git
	u = strings.TrimSuffix(u, ".git")
	// Take the last path component
	if idx := strings.LastIndex(u, "/"); idx >= 0 {
		u = u[idx+1:]
	}
	if u == "" {
		return "unknown"
	}
	return u
}

// HTTPConfig holds config for Docker-managed or external HTTP servers.
type HTTPConfig struct {
	Image       string            `json:"image,omitempty"`
	Port        int               `json:"port,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	HealthCheck string            `json:"health_check,omitempty"`
	URL         string            `json:"url,omitempty"` // for external HTTP servers
}

// RemoteConfig holds config for remote servers.
type RemoteConfig struct {
	URL  string     `json:"url"`
	Auth RemoteAuth `json:"auth"`
}

type RemoteAuth struct {
	Type         string `json:"type"` // "none", "private_url", "bearer", "api_key", "oauth"
	Token        string `json:"token,omitempty"`
	HeaderName   string `json:"header_name,omitempty"` // for api_key type
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	AuthURL      string `json:"auth_url,omitempty"`
	TokenURL     string `json:"token_url,omitempty"`
	Scopes       string `json:"scopes,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenExpiry  string `json:"token_expiry,omitempty"` // RFC3339
	// Registration tracking: stored so we can detect redirect_uri changes
	RegisteredRedirectURI string `json:"registered_redirect_uri,omitempty"`
	RegistrationEndpoint  string `json:"registration_endpoint,omitempty"`
}

type ServerStore struct {
	db     *DB
	crypto *ConfigEncryptor
}

func NewServerStore(db *DB, crypto *ConfigEncryptor) *ServerStore {
	return &ServerStore{db: db, crypto: crypto}
}

func (s *ServerStore) Create(srv *Server) error {
	if err := ValidateSlug(srv.Name); err != nil {
		return err
	}
	if srv.ID == "" {
		srv.ID = uuid.New().String()
	}
	srv.Status = StatusStopped
	srv.CreatedAt = time.Now()
	srv.UpdatedAt = time.Now()

	storedConfig, err := s.crypto.Encrypt(srv.Config)
	if err != nil {
		return fmt.Errorf("encrypting config: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO servers (id, name, display_name, server_type, config, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		srv.ID, srv.Name, srv.DisplayName, srv.ServerType, storedConfig, srv.Status, srv.CreatedAt, srv.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrSlugConflict
		}
		return fmt.Errorf("creating server: %w", err)
	}
	return nil
}

func (s *ServerStore) Get(id string) (*Server, error) {
	srv := &Server{}
	err := s.db.QueryRow(`
		SELECT id, name, display_name, server_type, config, status, COALESCE(error_msg, ''),
		       COALESCE(health, 'unknown'), health_check_at, COALESCE(health_error, ''),
		       optimize_enabled, created_at, updated_at
		FROM servers WHERE id = ?`, id,
	).Scan(&srv.ID, &srv.Name, &srv.DisplayName, &srv.ServerType, &srv.Config, &srv.Status, &srv.ErrorMsg,
		&srv.Health, &srv.HealthCheckAt, &srv.HealthError,
		&srv.OptimizeEnabled, &srv.CreatedAt, &srv.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting server: %w", err)
	}
	if plaintext, err := s.crypto.Decrypt(srv.Config); err != nil {
		return nil, fmt.Errorf("decrypting config: %w", err)
	} else {
		srv.Config = plaintext
	}
	return srv, nil
}

func (s *ServerStore) GetByName(name string) (*Server, error) {
	srv := &Server{}
	err := s.db.QueryRow(`
		SELECT id, name, display_name, server_type, config, status, COALESCE(error_msg, ''),
		       COALESCE(health, 'unknown'), health_check_at, COALESCE(health_error, ''),
		       optimize_enabled, created_at, updated_at
		FROM servers WHERE name = ?`, name,
	).Scan(&srv.ID, &srv.Name, &srv.DisplayName, &srv.ServerType, &srv.Config, &srv.Status, &srv.ErrorMsg,
		&srv.Health, &srv.HealthCheckAt, &srv.HealthError,
		&srv.OptimizeEnabled, &srv.CreatedAt, &srv.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting server by name: %w", err)
	}
	if plaintext, err := s.crypto.Decrypt(srv.Config); err != nil {
		return nil, fmt.Errorf("decrypting config: %w", err)
	} else {
		srv.Config = plaintext
	}
	return srv, nil
}

func (s *ServerStore) List() ([]*Server, error) {
	rows, err := s.db.Query(`
		SELECT id, name, display_name, server_type, config, status, COALESCE(error_msg, ''),
		       COALESCE(health, 'unknown'), health_check_at, COALESCE(health_error, ''),
		       optimize_enabled, created_at, updated_at
		FROM servers ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var servers []*Server
	for rows.Next() {
		srv := &Server{}
		if err := rows.Scan(&srv.ID, &srv.Name, &srv.DisplayName, &srv.ServerType, &srv.Config, &srv.Status, &srv.ErrorMsg,
			&srv.Health, &srv.HealthCheckAt, &srv.HealthError,
			&srv.OptimizeEnabled, &srv.CreatedAt, &srv.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning server: %w", err)
		}
		if plaintext, err := s.crypto.Decrypt(srv.Config); err != nil {
			return nil, fmt.Errorf("decrypting config for %s: %w", srv.Name, err)
		} else {
			srv.Config = plaintext
		}
		servers = append(servers, srv)
	}
	return servers, nil
}

func (s *ServerStore) Update(srv *Server) error {
	srv.UpdatedAt = time.Now()
	storedConfig, err := s.crypto.Encrypt(srv.Config)
	if err != nil {
		return fmt.Errorf("encrypting config: %w", err)
	}
	_, err = s.db.Exec(`
		UPDATE servers SET name = ?, display_name = ?, server_type = ?, config = ?, status = ?, error_msg = ?, updated_at = ?
		WHERE id = ?`,
		srv.Name, srv.DisplayName, srv.ServerType, storedConfig, srv.Status, srv.ErrorMsg, srv.UpdatedAt, srv.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrSlugConflict
		}
		return fmt.Errorf("updating server: %w", err)
	}
	return nil
}

func (s *ServerStore) UpdateStatus(id string, status ServerStatus, errMsg string) error {
	_, err := s.db.Exec(`
		UPDATE servers SET status = ?, error_msg = ?, updated_at = ? WHERE id = ?`,
		status, errMsg, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("updating server status: %w", err)
	}
	return nil
}

// UpdateHealth updates the MCP-level health check fields for a server.
func (s *ServerStore) UpdateHealth(id string, health HealthStatus, healthError string) error {
	now := time.Now()
	_, err := s.db.Exec(`
		UPDATE servers SET health = ?, health_check_at = ?, health_error = ?, updated_at = ? WHERE id = ?`,
		health, now, healthError, now, id,
	)
	if err != nil {
		return fmt.Errorf("updating server health: %w", err)
	}
	return nil
}

// UpdateConfig updates only the JSON config column for a server.
func (s *ServerStore) UpdateConfig(id string, config json.RawMessage) error {
	storedConfig, err := s.crypto.Encrypt(config)
	if err != nil {
		return fmt.Errorf("encrypting config: %w", err)
	}
	_, err = s.db.Exec(`UPDATE servers SET config = ?, updated_at = ? WHERE id = ?`,
		storedConfig, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("updating server config: %w", err)
	}
	return nil
}

// SetOptimizeEnabled toggles the optimize_enabled flag for a server.
func (s *ServerStore) SetOptimizeEnabled(id string, enabled bool) error {
	_, err := s.db.Exec("UPDATE servers SET optimize_enabled = ?, updated_at = ? WHERE id = ?",
		enabled, time.Now(), id)
	return err
}

func (s *ServerStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM servers WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting server: %w", err)
	}
	return nil
}
