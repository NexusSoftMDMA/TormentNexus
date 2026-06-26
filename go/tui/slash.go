package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/NexusSoftMDMA/TormentNexus/agents"
	"github.com/NexusSoftMDMA/TormentNexus/foundation/adapters"
	foundationorchestration "github.com/NexusSoftMDMA/TormentNexus/foundation/orchestration"
	foundationrepomap "github.com/NexusSoftMDMA/TormentNexus/foundation/repomap"
)

// ProcessSlashCommand mimics Claude Code's native terminal interception primitives.
// If an input begins with '/', it bypasses the LLM and executes directly as a local primitive.
func ProcessSlashCommand(cmd string, m *model) (tea.Model, tea.Cmd) {
	cmd = strings.TrimSpace(cmd)
	parts := strings.Split(cmd, " ")

	switch parts[0] {
	case "/help":
		return handleHelp(m)
	case "/clear":
		return handleClear(m)
	case "/commit":
		return handleCommit(m)
	case "/plan":
		return handlePlan(m, strings.TrimSpace(strings.TrimPrefix(cmd, "/plan")))
	case "/repomap":
		return handleRepoMap(m)
	case "/providers":
		return handleProviders(m)
	case "/adapters":
		return handleAdapters(m)
	case "/mcp", "/mcptools":
		return handleMCPTools(m)
	case "/exit", "/quit":
		return *m, tea.Quit
	default:
		return handleUnknown(m, parts[0])
	}
}

func handleHelp(m *model) (tea.Model, tea.Cmd) {
	m.loading = false
	// Simulate glamour markdown output locally bypassing the Director
	m.history = append(m.history, `[System] Parity Slash Commands:
  /help      - This menu
  /clear     - Wipes contextual agent memory
  /commit    - Autonomously generates a standard Git commit
  /plan      - Build a foundation-backed orchestration plan
  /repomap   - Generate a foundation-backed repo map
  /providers - Show provider visibility and defaults
  /adapters  - Show TormentNexus/TormentNexus + MCP adapter status
  /mcp       - Show adapter-backed MCP tool hints
  /exit      - Closes tormentnexus`)
	return *m, nil
}

func handleClear(m *model) (tea.Model, tea.Cmd) {
	m.loading = false
	workingDir := m.director.WorkingDir
	m.history = []string{
		"[System] Memory crystal wiped. Context reset to null.",
	}
	m.director = agents.NewDirector(&agents.DefaultProvider{})
	m.director.WorkingDir = workingDir
	return *m, nil
}

func handleCommit(m *model) (tea.Model, tea.Cmd) {
	m.loading = false
	m.history = append(m.history, "[System] Extracting git diff to generate native algorithmic commit message...")
	// We'd tap native.go tools here to run 'git diff' and pass it to Director exclusively for formatting
	return *m, nil
}

func handlePlan(m *model, prompt string) (tea.Model, tea.Cmd) {
	m.loading = false
	if strings.TrimSpace(prompt) == "" {
		m.history = append(m.history, "[Error] /plan requires a prompt")
		return *m, nil
	}
	plan, err := foundationorchestration.BuildPlan(foundationorchestration.PlanRequest{Prompt: prompt, WorkingDir: m.director.WorkingDir, IncludeRepo: true, MaxRepoFiles: 8})
	if err != nil {
		m.history = append(m.history, fmt.Sprintf("[Error] plan generation failed: %v", err))
		return *m, nil
	}
	payload, _ := json.MarshalIndent(plan, "", "  ")
	m.history = append(m.history, "[Foundation Plan]\n"+string(payload))
	return *m, nil
}

func handleRepoMap(m *model) (tea.Model, tea.Cmd) {
	m.loading = false
	result, err := foundationrepomap.Generate(foundationrepomap.Options{BaseDir: m.director.WorkingDir, MaxFiles: 12})
	if err != nil {
		m.history = append(m.history, fmt.Sprintf("[Error] repomap generation failed: %v", err))
		return *m, nil
	}
	m.history = append(m.history, "[Foundation RepoMap]\n"+result.Map)
	return *m, nil
}

func handleProviders(m *model) (tea.Model, tea.Cmd) {
	m.loading = false
	payload, _ := json.MarshalIndent(adapters.BuildProviderStatus(), "", "  ")
	m.history = append(m.history, "[Foundation Providers]\n"+string(payload))
	return *m, nil
}

func handleAdapters(m *model) (tea.Model, tea.Cmd) {
	m.loading = false
	hyper := adapters.NewTormentNexusAdapter(m.director.WorkingDir)
	mcpAdapter := adapters.NewMCPAdapter(m.director.WorkingDir)
	payload, _ := json.MarshalIndent(map[string]any{
		"tormentnexus": hyper.Status(),
		"mcp":       mcpAdapter.Status(),
	}, "", "  ")
	m.history = append(m.history, "[Foundation Adapters]\n"+string(payload))
	return *m, nil
}

func handleMCPTools(m *model) (tea.Model, tea.Cmd) {
	m.loading = false
	mcpAdapter := adapters.NewMCPAdapter(m.director.WorkingDir)
	tools, err := mcpAdapter.ListTools()
	if err != nil {
		m.history = append(m.history, fmt.Sprintf("[Error] mcp tools unavailable: %v", err))
		return *m, nil
	}
	payload, _ := json.MarshalIndent(map[string]any{"tools": tools}, "", "  ")
	m.history = append(m.history, "[Foundation MCP Tools]\n"+string(payload))
	return *m, nil
}

func handleUnknown(m *model, cmd string) (tea.Model, tea.Cmd) {
	m.loading = false
	m.history = append(m.history, fmt.Sprintf("[Error] Unknown primitive: %s", cmd))
	return *m, nil
}
