package mcp

import "github.com/NexusSoftMDMA/TormentNexus/foundation/adapters"

// Config binds the ~/.tormentnexus/mcp.json native parsing.
type Config struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

type ServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// ReadScopedClient strictly constructs OS subprocess environment lists without pulling globally.
func ParseMetadataContext() (*Config, error) {
	_, conf, err := adapters.ParseMCPConfig("")
	if err != nil {
		return nil, err
	}
	converted := &Config{MCPServers: map[string]ServerConfig{}}
	for name, server := range conf.MCPServers {
		converted.MCPServers[name] = ServerConfig(server)
	}
	return converted, nil
}

// FlattenEnv constructs the isolated subprocess environmental bindings.
func (s *ServerConfig) FlattenEnv() []string {
	server := adapters.MCPServerConfig(*s)
	return server.FlattenEnv()
}
