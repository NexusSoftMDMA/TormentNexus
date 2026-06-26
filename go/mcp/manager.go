package mcp

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/NexusSoftMDMA/TormentNexus/foundation/adapters"
)

// ServerManager handles the installation and lifecycle of MCP servers (Smithery parity).
type ServerManager struct {
	RegistryPath string
	adapter      *adapters.MCPAdapter
}

func NewServerManager() *ServerManager {
	cwd, _ := os.Getwd()
	adapter := adapters.NewMCPAdapter(cwd)
	status := adapter.Status()
	registryPath := "./.supercli/mcp_servers"
	if status.ConfigPath != "" {
		registryPath = status.ConfigPath
	}
	return &ServerManager{
		RegistryPath: registryPath,
		adapter:      adapter,
	}
}

// Install from an npm package (e.g., npx @smithery/cli install).
func (sm *ServerManager) InstallNPXServer(packageName string) error {
	fmt.Printf("Installing MCP server via NPX: %s\n", packageName)
	cmd := exec.Command("npm", "install", "-g", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s: %s\n%s", packageName, err, output)
	}
	return nil
}

// StartServer launches an MCP server process.
func (sm *ServerManager) StartServer(command string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

// StartConfiguredServer launches a configured MCP server by name via the adapter seam.
func (sm *ServerManager) StartConfiguredServer(name string) (*exec.Cmd, error) {
	return sm.adapter.StartConfiguredServer(name)
}

// ListConfiguredTools returns adapter-backed MCP tool hints.
func (sm *ServerManager) ListConfiguredTools() ([]string, error) {
	return sm.adapter.ListTools()
}

// RouteConfiguredToolCall returns the mediated route for a configured MCP tool call.
func (sm *ServerManager) RouteConfiguredToolCall(serverName, toolName string, args map[string]interface{}) (string, error) {
	result, err := sm.adapter.CallTool(adapters.MCPCallRequest{
		ServerName: serverName,
		ToolName:   toolName,
		Arguments:  args,
	})
	if err != nil {
		return "", err
	}
	return result.Route, nil
}
