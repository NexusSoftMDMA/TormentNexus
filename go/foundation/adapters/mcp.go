package adapters

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/NexusSoftMDMA/TormentNexus/tormentnexus"
)

type MCPStatus struct {
	ConfigPath string            `json:"configPath,omitempty"`
	Servers    []MCPServerStatus `json:"servers,omitempty"`
	Warnings   []string          `json:"warnings,omitempty"`
}

type MCPServerStatus struct {
	Name       string   `json:"name"`
	Command    string   `json:"command,omitempty"`
	Args       []string `json:"args,omitempty"`
	HasEnv     bool     `json:"hasEnv"`
	ToolHints  []string `json:"toolHints,omitempty"`
	RouteHint  string   `json:"routeHint,omitempty"`
	Executable bool     `json:"executable"`
}

type MCPCallRequest struct {
	ServerName string                 `json:"serverName"`
	ToolName   string                 `json:"toolName"`
	Arguments  map[string]interface{} `json:"arguments,omitempty"`
}

type MCPCallResult struct {
	ServerName string                 `json:"serverName"`
	ToolName   string                 `json:"toolName"`
	Arguments  map[string]interface{} `json:"arguments,omitempty"`
	Route      string                 `json:"route"`
	Summary    string                 `json:"summary"`
	Executable bool                   `json:"executable"`
}

type MCPAdapter struct {
	tormentnexusAdapter *tormentnexus.Adapter
	workingDir  string
	homeDir     string
}

func NewMCPAdapter(workingDir string) *MCPAdapter {
	homeDir, _ := os.UserHomeDir()
	return &MCPAdapter{
		tormentnexusAdapter: tormentnexus.NewAdapter(),
		workingDir:  workingDir,
		homeDir:     homeDir,
	}
}

func (a *MCPAdapter) Status() MCPStatus {
	configPath, conf, err := ParseMCPConfig(a.homeDir)
	status := MCPStatus{ConfigPath: configPath}
	if err != nil {
		status.Warnings = append(status.Warnings, err.Error())
		return status
	}
	names := make([]string, 0, len(conf.MCPServers))
	for name := range conf.MCPServers {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		server := conf.MCPServers[name]
		status.Servers = append(status.Servers, MCPServerStatus{
			Name:       name,
			Command:    server.Command,
			Args:       append([]string(nil), server.Args...),
			HasEnv:     len(server.Env) > 0,
			ToolHints:  defaultToolHintsForServer(name, server),
			RouteHint:  a.routeHint(name),
			Executable: commandResolvable(server.Command),
		})
	}
	return status
}

func (a *MCPAdapter) ListTools() ([]string, error) {
	status := a.Status()
	if len(status.Servers) == 0 {
		if len(status.Warnings) > 0 {
			return nil, fmt.Errorf("%s", strings.Join(status.Warnings, "; "))
		}
		return nil, fmt.Errorf("no configured MCP servers")
	}
	tools := make([]string, 0)
	for _, server := range status.Servers {
		tools = append(tools, server.ToolHints...)
	}
	sort.Strings(tools)
	return tools, nil
}

func (a *MCPAdapter) RouteCall(serverName, request string) string {
	payload := fmt.Sprintf("%s:%s", strings.TrimSpace(serverName), strings.TrimSpace(request))
	if a.tormentnexusAdapter == nil {
		return payload
	}
	return a.tormentnexusAdapter.RouteMCP(payload)
}

func (a *MCPAdapter) CallTool(req MCPCallRequest) (MCPCallResult, error) {
	if strings.TrimSpace(req.ServerName) == "" {
		return MCPCallResult{}, fmt.Errorf("server name is required")
	}
	if strings.TrimSpace(req.ToolName) == "" {
		return MCPCallResult{}, fmt.Errorf("tool name is required")
	}
	server, ok := a.LookupServer(req.ServerName)
	if !ok {
		return MCPCallResult{}, fmt.Errorf("unknown MCP server: %s", req.ServerName)
	}
	result := MCPCallResult{
		ServerName: req.ServerName,
		ToolName:   req.ToolName,
		Arguments:  req.Arguments,
		Route:      a.RouteCall(req.ServerName, req.ToolName),
		Summary:    fmt.Sprintf("Prepared MCP tool call %s on %s", req.ToolName, req.ServerName),
		Executable: commandResolvable(server.Command),
	}
	return result, nil
}

func (a *MCPAdapter) LookupServer(name string) (MCPServerConfig, bool) {
	_, conf, err := ParseMCPConfig(a.homeDir)
	if err != nil {
		return MCPServerConfig{}, false
	}
	server, ok := conf.MCPServers[name]
	return server, ok
}

func (a *MCPAdapter) StartConfiguredServer(name string) (*exec.Cmd, error) {
	server, ok := a.LookupServer(name)
	if !ok {
		return nil, fmt.Errorf("unknown MCP server: %s", name)
	}
	if strings.TrimSpace(server.Command) == "" {
		return nil, fmt.Errorf("MCP server %s has no command", name)
	}
	cmd := exec.Command(server.Command, server.Args...)
	cmd.Env = server.FlattenEnv()
	if a.workingDir != "" {
		cmd.Dir = a.workingDir
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func defaultToolHintsForServer(name string, server MCPServerConfig) []string {
	base := strings.ToLower(strings.TrimSpace(name))
	if base == "" {
		base = "mcp"
	}
	return []string{
		fmt.Sprintf("mcp:%s:list-tools", base),
		fmt.Sprintf("mcp:%s:call-tool", base),
	}
}

func (a *MCPAdapter) routeHint(name string) string {
	if a.tormentnexusAdapter == nil {
		return name
	}
	return a.tormentnexusAdapter.RouteMCP(name)
}

func commandResolvable(command string) bool {
	if strings.TrimSpace(command) == "" {
		return false
	}
	_, err := exec.LookPath(command)
	return err == nil
}
