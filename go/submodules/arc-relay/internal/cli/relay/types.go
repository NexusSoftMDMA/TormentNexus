package relay

import "encoding/json"

// CreateServerRequest is the payload for creating a server on the relay.
// Mirrors the relay's store.Server model.
type CreateServerRequest struct {
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	ServerType  string          `json:"server_type"` // "stdio", "http", "remote"
	Config      json.RawMessage `json:"config"`
}

// StdioConfig is the config for a Docker-managed stdio MCP server.
type StdioConfig struct {
	Image      string            `json:"image,omitempty"`
	Build      *StdioBuildConfig `json:"build,omitempty"`
	Entrypoint []string          `json:"entrypoint,omitempty"`
	Command    []string          `json:"command,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
}

// StdioBuildConfig describes how to auto-build a Docker image from a package.
type StdioBuildConfig struct {
	Runtime string `json:"runtime"`           // "python" or "node"
	Package string `json:"package"`           // pip/npm package name
	Version string `json:"version,omitempty"` // package version (empty = latest)
	GitURL  string `json:"git_url,omitempty"` // alternative: build from git repo
}

// HTTPConfig is the config for a Docker-managed or external HTTP MCP server.
type HTTPConfig struct {
	Image       string            `json:"image,omitempty"`
	Port        int               `json:"port,omitempty"`
	URL         string            `json:"url,omitempty"`
	HealthCheck string            `json:"health_check,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
}

// RemoteConfig is the config for a remote MCP server.
type RemoteConfig struct {
	URL  string     `json:"url"`
	Auth RemoteAuth `json:"auth"`
}

// RemoteAuth describes authentication for a remote server.
type RemoteAuth struct {
	Type       string `json:"type"`                  // "none", "bearer", "api_key"
	Token      string `json:"token,omitempty"`       // for bearer type
	HeaderName string `json:"header_name,omitempty"` // for api_key type
}

// ServerDetail is the full server response from the relay API,
// including the config blob.
type ServerDetail struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	ServerType  string          `json:"server_type"`
	Config      json.RawMessage `json:"config"`
	Status      string          `json:"status"`
	ErrorMsg    string          `json:"error_msg,omitempty"`
	CreatedAt   string          `json:"created_at"`
}
