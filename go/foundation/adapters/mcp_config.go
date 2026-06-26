package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

func ParseMCPConfig(homeDir string) (string, *MCPConfig, error) {
	if strings.TrimSpace(homeDir) == "" {
		homeDir, _ = os.UserHomeDir()
	}
	path := filepath.Join(homeDir, ".tormentnexus", "mcp.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return path, nil, fmt.Errorf("missing .tormentnexus/mcp.json definition: %w", err)
	}
	var conf MCPConfig
	if err := json.Unmarshal(data, &conf); err != nil {
		return path, nil, err
	}
	if conf.MCPServers == nil {
		conf.MCPServers = map[string]MCPServerConfig{}
	}
	return path, &conf, nil
}

func (s *MCPServerConfig) FlattenEnv() []string {
	var envList []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "PATH=") || strings.HasPrefix(e, "NODE_ENV=") {
			envList = append(envList, e)
		}
	}
	for k, v := range s.Env {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}
	return envList
}
