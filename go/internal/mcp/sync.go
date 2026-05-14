package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type SupportedClient string

const (
	ClientClaudeDesktop SupportedClient = "claude-desktop"
	ClientCursor        SupportedClient = "cursor"
	ClientVSCode        SupportedClient = "vscode"
)

var SupportedClients = []SupportedClient{
	ClientClaudeDesktop,
	ClientCursor,
	ClientVSCode,
}

type ResolvedTarget struct {
	Client     SupportedClient `json:"client"`
	Path       string          `json:"path"`
	Candidates []string        `json:"candidates"`
	Exists     bool            `json:"exists"`
}

type SyncResult struct {
	Client      SupportedClient `json:"client"`
	TargetPath  string          `json:"targetPath"`
	Existed     bool            `json:"existed"`
	ServerCount int             `json:"serverCount"`
	Document    interface{}     `json:"document,omitempty"`
	JSON        string          `json:"json,omitempty"`
	Written     bool            `json:"written"`
}

func ResolveClientTargets(homeDir string, appData string, cwd string) []ResolvedTarget {
	var results []ResolvedTarget
	for _, client := range SupportedClients {
		results = append(results, ResolveClientTarget(client, "", homeDir, appData, cwd))
	}
	return results
}

func ResolveClientTarget(client SupportedClient, overridePath string, homeDir string, appData string, cwd string) ResolvedTarget {
	candidates := []string{}
	if overridePath != "" {
		candidates = []string{overridePath}
	} else {
		candidates = getClientCandidates(client, homeDir, appData, cwd)
	}

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

	return ResolvedTarget{
		Client:     client,
		Path:       existingPath,
		Candidates: candidates,
		Exists:     exists,
	}
}

func getClientCandidates(client SupportedClient, homeDir string, appData string, cwd string) []string {
	if appData == "" {
		appData = filepath.Join(homeDir, "AppData", "Roaming")
	}

	switch client {
	case ClientClaudeDesktop:
		return byPlatform([]string{filepath.Join(appData, "Claude", "claude_desktop_config.json")},
			[]string{filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")},
			[]string{filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json")})
	case ClientCursor:
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
	case ClientVSCode:
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

func PreviewClientConfig(client SupportedClient, target ResolvedTarget, servers map[string]McpServerConfig) (*SyncResult, error) {
	// 1. Read existing config
	document := make(map[string]interface{})
	if data, err := os.ReadFile(target.Path); err == nil {
		_ = json.Unmarshal(data, &document)
	}

	// 2. Prepare new mcpServers block
	mcpServers := make(map[string]interface{})
	for name, cfg := range servers {
		if cfg.Command != "" {
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
			mcpServers[name] = map[string]interface{}{
				"url": cfg.URL,
			}
		}
	}

	// 3. Merge
	document["mcpServers"] = mcpServers

	// 4. Generate JSON string
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, err
	}

	return &SyncResult{
		Client:      client,
		TargetPath:  target.Path,
		Existed:     target.Exists,
		ServerCount: len(mcpServers),
		Document:    document,
		JSON:        string(data) + "\n",
		Written:     false,
	}, nil
}

func WriteClientConfig(preview *SyncResult) error {
	if err := os.MkdirAll(filepath.Dir(preview.TargetPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(preview.TargetPath, []byte(preview.JSON), 0644)
}
