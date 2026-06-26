package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type SupportedClient string

const (
	ClaudeDesktop SupportedClient = "claude-desktop"
	Cursor        SupportedClient = "cursor"
	VSCode        SupportedClient = "vscode"
)

type ResolvedTarget struct {
	Client     SupportedClient `json:"client"`
	Path       string          `json:"path"`
	Candidates []string        `json:"candidates"`
	Exists     bool            `json:"exists"`
}

type SyncResult struct {
	Client      SupportedClient `json:"client"`
	TargetPath  string          `json:"targetPath"`
	ServerCount int             `json:"serverCount"`
	Written     bool            `json:"written"`
}

func ResolveClientTargets(homeDir string, appData string, cwd string) []ResolvedTarget {
	clients := []SupportedClient{ClaudeDesktop, Cursor, VSCode}
	var results []ResolvedTarget

	for _, client := range clients {
		candidates := getClientCandidates(client, homeDir, appData, cwd)
		var existingPath string
		exists := false

		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				existingPath = c
				exists = true
				break
			}
		}

		if !exists && len(candidates) > 0 {
			existingPath = candidates[0]
		}

		results = append(results, ResolvedTarget{
			Client:     client,
			Path:       existingPath,
			Candidates: candidates,
			Exists:     exists,
		})
	}

	return results
}

func getClientCandidates(client SupportedClient, homeDir string, appData string, cwd string) []string {
	if appData == "" {
		if runtime.GOOS == "windows" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
	}

	switch client {
	case ClaudeDesktop:
		return byPlatform([]string{filepath.Join(appData, "Claude", "claude_desktop_config.json")},
			[]string{filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")},
			[]string{filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json")})
	case Cursor:
		return byPlatform([]string{
			filepath.Join(appData, "Cursor", "User", "globalStorage", "mcp-servers.json"),
			filepath.Join(appData, "Cursor", "User", "mcp.json"),
		}, []string{
			filepath.Join(homeDir, "Library", "Application Support", "Cursor", "User", "globalStorage", "mcp-servers.json"),
			filepath.Join(homeDir, "Library", "Application Support", "Cursor", "User", "mcp.json"),
		}, []string{
			filepath.Join(homeDir, ".config", "Cursor", "User", "globalStorage", "mcp-servers.json"),
			filepath.Join(homeDir, ".config", "Cursor", "User", "mcp.json"),
		})
	case VSCode:
		return byPlatform([]string{
			filepath.Join(appData, "Code", "User", "globalStorage", "mcp-servers.json"),
			filepath.Join(appData, "Code", "User", "settings.json"),
			filepath.Join(cwd, ".vscode", "mcp.json"),
		}, []string{
			filepath.Join(homeDir, "Library", "Application Support", "Code", "User", "globalStorage", "mcp-servers.json"),
			filepath.Join(homeDir, "Library", "Application Support", "Code", "User", "settings.json"),
			filepath.Join(cwd, ".vscode", "mcp.json"),
		}, []string{
			filepath.Join(homeDir, ".config", "Code", "User", "globalStorage", "mcp-servers.json"),
			filepath.Join(homeDir, ".config", "Code", "User", "settings.json"),
			filepath.Join(cwd, ".vscode", "mcp.json"),
		})
	}
	return nil
}

func byPlatform(win, mac, linux []string) []string {
	switch runtime.GOOS {
	case "windows":
		return win
	case "darwin":
		return mac
	default:
		return linux
	}
}

func SyncToClient(client SupportedClient, targetPath string, servers map[string]McpServerConfig) (*SyncResult, error) {
	// 1. Read existing config
	existing := make(map[string]interface{})
	if data, err := os.ReadFile(targetPath); err == nil {
		_ = json.Unmarshal(data, &existing)
	}

	// 2. Prepare new mcpServers block
	mcpServers := make(map[string]interface{})
	for name, cfg := range servers {
		if cfg.Command != "" {
			// Stdio Server
			def := map[string]interface{}{
				"command": cfg.Command,
			}
			if len(cfg.Args) > 0 {
				def["args"] = cfg.Args
			}
			if len(cfg.Env) > 0 {
				def["env"] = cfg.Env
			}
			mcpServers[name] = def
		} else if cfg.URL != "" {
			// HTTP/SSE Server
			def := map[string]interface{}{
				"url": cfg.URL,
			}

			headers := make(map[string]string)
			for k, v := range cfg.Headers {
				headers[k] = v
			}
			if cfg.BearerToken != "" {
				headers["Authorization"] = "Bearer " + cfg.BearerToken
			}

			if len(headers) > 0 {
				def["headers"] = headers
			}
			mcpServers[name] = def
		}
	}

	// 3. Merge or Replace
	// Some tools like Cursor store MCP servers in a sub-property, others top-level
	if client == VSCode && strings.HasSuffix(targetPath, "settings.json") {
		existing["mcp.servers"] = mcpServers
	} else {
		existing["mcpServers"] = mcpServers
	}

	// 4. Write
	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return nil, err
	}

	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return nil, err
	}

	return &SyncResult{
		Client:      client,
		TargetPath:  targetPath,
		ServerCount: len(mcpServers),
		Written:     true,
	}, nil
}
