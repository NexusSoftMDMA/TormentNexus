package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/NexusSoftMDMA/TormentNexus/foundation/compat"
	foundationpi "github.com/NexusSoftMDMA/TormentNexus/foundation/pi"
	foundationrepomap "github.com/NexusSoftMDMA/TormentNexus/foundation/repomap"
)

// Registry holds all available tools for the agent.
type Registry struct {
	Tools []Tool
}

// Tool describes a model-facing callable tool surface.
type Tool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
	Execute     func(args map[string]interface{}) (string, error)
}

func NewRegistry() *Registry {
	r := &Registry{}
	r.registerFoundationTools()
	r.registerCoreTools()
	r.registerFileTools()
	r.registerRepoMapTools()
	r.registerInterpreterTools()
	r.registerSearchTools()
	r.registerAdvancedTools()
	r.registerRefactoringTools()
	r.registerCloudTools()
	r.registerIntegrationsTools()
	r.registerBookmarkTools()
	r.registerGUITools()
	r.registerCloudOrchestratorTools()
	r.registerSystemTools()
	r.registerLlamafileTools()
	return r
}

func (r *Registry) registerFoundationTools() {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	runtime := foundationpi.NewRuntime(cwd, nil)
	catalog := compat.DefaultCatalog()
	for _, contract := range catalog.ContractsBySource("pi") {
		contract := contract
		r.Tools = append(r.Tools, Tool{
			Name:        contract.Name,
			Description: contract.Description,
			Parameters:  append(json.RawMessage(nil), contract.Parameters...),
			Execute: func(args map[string]interface{}) (string, error) {
				raw, err := json.Marshal(args)
				if err != nil {
					return "", fmt.Errorf("marshal args for %s: %w", contract.Name, err)
				}
				result, execErr := runtime.ExecuteTool(context.Background(), "", contract.Name, raw, nil)
				formatted := formatFoundationToolResult(result)
				if execErr != nil {
					if formatted != "" {
						return formatted, execErr
					}
					return "", execErr
				}
				return formatted, nil
			},
		})
	}
}

func (r *Registry) registerCoreTools() {
	r.Tools = append(r.Tools, Tool{
		Name:        "run_shell_command",
		Description: "Legacy shell execution alias. Prefer bash for exact Pi-compatible tool naming.",
		Parameters:  json.RawMessage(`{"type":"object","required":["command"],"properties":{"command":{"type":"string"}},"additionalProperties":false}`),
		Execute: func(args map[string]interface{}) (string, error) {
			cmdStr, _ := args["command"].(string)
			if strings.TrimSpace(cmdStr) == "" {
				return "", fmt.Errorf("command must be a string")
			}
			cmd := shellCommand(cmdStr)
			out, err := cmd.CombinedOutput()
			return string(out), err
		},
	})
}

func (r *Registry) registerSearchTools() {
	r.Tools = append(r.Tools, Tool{
		Name:        "search",
		Description: "Recursively searches for a pattern in the codebase. Arguments: pattern (string)",
		Parameters:  json.RawMessage(`{"type":"object","required":["pattern"],"properties":{"pattern":{"type":"string"}},"additionalProperties":false}`),
		Execute: func(args map[string]interface{}) (string, error) {
			pattern, _ := args["pattern"].(string)
			var results string
			filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() && (info.Name() == ".git" || info.Name() == "node_modules") {
					return filepath.SkipDir
				}
				if !info.IsDir() {
					content, _ := os.ReadFile(path)
					if strings.Contains(string(content), pattern) {
						rel, _ := filepath.Rel(".", path)
						results += fmt.Sprintf("Match in %s\n", rel)
					}
				}
				return nil
			})
			return results + "Search functionality complete", nil
		},
	})
}

func (r *Registry) registerRepoMapTools() {
	r.Tools = append(r.Tools, Tool{
		Name:        "repomap",
		Description: "Generate a ranked repository map with lightweight symbol summaries.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"dir":{"type":"string"},"mention_file":{"type":"array","items":{"type":"string"}},"mention_ident":{"type":"array","items":{"type":"string"}},"max_files":{"type":"integer"},"include_tests":{"type":"boolean"}},"additionalProperties":false}`),
		Execute: func(args map[string]interface{}) (string, error) {
			baseDir, _ := args["dir"].(string)
			if strings.TrimSpace(baseDir) == "" {
				baseDir = "."
			}
			result, err := foundationrepomap.Generate(foundationrepomap.Options{
				BaseDir:         baseDir,
				MentionedFiles:  toStringSlice(args["mention_file"]),
				MentionedIdents: toStringSlice(args["mention_ident"]),
				MaxFiles:        toInt(args["max_files"], 40),
				IncludeTests:    toBool(args["include_tests"]),
			})
			if err != nil {
				return "", err
			}
			return result.Map, nil
		},
	})
}

func formatFoundationToolResult(result *foundationpi.ToolResult) string {
	if result == nil {
		return ""
	}
	textBlocks := make([]string, 0, len(result.Content))
	textOnly := true
	for _, block := range result.Content {
		switch value := block.(type) {
		case foundationpi.TextContent:
			textBlocks = append(textBlocks, value.Text)
		case map[string]interface{}:
			if text, _ := value["text"].(string); text != "" {
				textBlocks = append(textBlocks, text)
			} else {
				textOnly = false
			}
		default:
			textOnly = false
		}
	}
	if textOnly && result.Details == nil {
		return strings.Join(textBlocks, "\n")
	}
	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return strings.Join(textBlocks, "\n")
	}
	return string(payload)
}

func shellCommand(command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", command)
	}
	return exec.Command("sh", "-lc", command)
}

func toStringSlice(value interface{}) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []interface{}:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if str, ok := item.(string); ok && strings.TrimSpace(str) != "" {
				out = append(out, str)
			}
		}
		return out
	default:
		return nil
	}
}

func toInt(value interface{}, fallback int) int {
	switch typed := value.(type) {
	case int:
		return typed
	case float64:
		return int(typed)
	default:
		return fallback
	}
}

func toBool(value interface{}) bool {
	if typed, ok := value.(bool); ok {
		return typed
	}
	return false
}
