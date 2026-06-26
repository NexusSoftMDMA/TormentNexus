package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// lspServer describes a language server configuration.
type lspServer struct {
	Name            string   `json:"name"`
	Extensions      []string `json:"extensions"`
	Command         []string `json:"command"`
	RootDir         string   `json:"rootDir"`
	RestartInterval int      `json:"restartInterval,omitempty"`
}

// defaultServers are the built‑in language server configurations.
var defaultServers = []lspServer{
	{
		Name:       "typescript-language-server",
		Extensions: []string{"js", "ts", "jsx", "tsx"},
		Command:    []string{"npx", "--", "typescript-language-server", "--stdio"},
		RootDir:    ".",
	},
	{
		Name:       "python-lsp-server",
		Extensions: []string{"py", "pyi"},
		Command:    []string{"uvx", "--from", "python-lsp-server", "pylsp"},
		RootDir:    ".",
		RestartInterval: 5,
	},
	{
		Name:       "gopls",
		Extensions: []string{"go"},
		Command:    []string{"gopls"},
		RootDir:    ".",
	},
}

// findServersByExtension returns all servers that handle the given extension.
func findServersByExtension(ext string) []lspServer {
	var result []lspServer
	for _, s := range defaultServers {
		for _, e := range s.Extensions {
			if strings.EqualFold(ext, e) {
				result = append(result, s)
				break
			}
		}
	}
	return result
}

// HandleListServers lists all available language servers.
func HandleListServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data, marshalErr := json.MarshalIndent(defaultServers, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal servers: %v", marshalErr))
}

	return ok(string(data))
}

// HandleGetServer returns details about a specific server by index (0‑based).
func HandleGetServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	idx, _ :=getInt(args, "index")
	if idx < 0 || idx >= len(defaultServers) {
		return err(fmt.Sprintf("invalid server index %d (0–%d)", idx, len(defaultServers)-1))
}

	server := defaultServers[idx]
	data, marshalErr := json.MarshalIndent(server, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal server: %v", marshalErr))
}

	return ok(string(data))
}

// HandleCheckFile returns which language server can handle a given file.
func HandleCheckFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filename, _ :=getString(args, "filename")
	if filename == "" {
		return err("filename argument is required")
}

	ext := strings.TrimPrefix(filepath.Ext(filename), ".")
	if ext == "" {
		return err("file has no extension, cannot determine language server")
}

	servers := findServersByExtension(ext)
	if len(servers) == 0 {
		return ok(fmt.Sprintf("No language server configured for .%s files", ext))
}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("File %q (extension .%s) can be handled by:\n", filename, ext))
	for _, s := range servers {
		sb.WriteString(fmt.Sprintf("  - %s (%s)\n", s.Name, strings.Join(s.Command, " ")))

	return ok(sb.String())
}
}