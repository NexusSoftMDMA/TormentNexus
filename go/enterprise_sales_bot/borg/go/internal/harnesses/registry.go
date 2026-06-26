package harnesses

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

type Definition struct {
	ID                  string   `json:"id"`
	Description         string   `json:"description"`
	Maturity            string   `json:"maturity"`
	Primary             bool     `json:"primary,omitempty"`
	SubmodulePath       string   `json:"submodulePath,omitempty"`
	Upstream            string   `json:"upstream,omitempty"`
	Runtime             string   `json:"runtime,omitempty"`
	LaunchCommand       string   `json:"launchCommand,omitempty"`
	Capabilities        []string `json:"capabilities,omitempty"`
	ParityNotes         string   `json:"parityNotes,omitempty"`
	ToolCallCount       int      `json:"toolCallCount,omitempty"`
	ToolCallNames       []string `json:"toolCallNames,omitempty"`
	ToolSource          string   `json:"toolSource,omitempty"`
	ToolInventoryStatus string   `json:"toolInventoryStatus"`
	IntegrationLevel    string   `json:"integrationLevel"`
	Installed           bool     `json:"installed"`
}

type Summary struct {
	SourceBackedHarnessCount    int `json:"sourceBackedHarnessCount"`
	MetadataOnlyHarnessCount    int `json:"metadataOnlyHarnessCount"`
	OperatorDefinedHarnessCount int `json:"operatorDefinedHarnessCount"`
	SourceBackedToolCount       int `json:"sourceBackedToolCount"`
}

func List(workspaceRoot string, tools []controlplane.Tool) []Definition {
	availableTools := make(map[string]bool, len(tools))
	for _, tool := range tools {
		if tool.Available {
			availableTools[tool.Type] = true
		}
	}

	tormentnexusTools := tormentnexusToolNames(workspaceRoot)
	externalHarnessNote := "External harness; tormentnexus currently tracks install/runtime metadata only, not a source-backed tool registry."
	metadataHarness := func(id, description, maturity, runtime string) Definition {
		return Definition{
			ID:                  id,
			Description:         description,
			Maturity:            maturity,
			Runtime:             runtime,
			ParityNotes:         externalHarnessNote,
			ToolInventoryStatus: "metadata-only",
			IntegrationLevel:    "metadata-only",
			Installed:           availableTools[id],
		}
	}
	definitions := []Definition{
		{
			ID:                  "tormentnexus",
			Description:         "tormentnexus Go CLI harness",
			Maturity:            "Experimental",
			Primary:             true,
			SubmodulePath:       "submodules/tormentnexus",
			Upstream:            "https://github.com/NexusSoftMDMA/TormentNexus",
			Runtime:             "Go / Cobra / TUI",
			LaunchCommand:       "go run .",
			Capabilities:        []string{"repl", "pipe", "tormentnexus-adapter", "tool-registry"},
			ParityNotes:         "tormentnexus can read tormentnexus tool calls directly from the assimilated submodule source.",
			ToolCallCount:       len(tormentnexusTools),
			ToolCallNames:       tormentnexusTools,
			ToolSource:          "submodules/tormentnexus/tools/*.go",
			ToolInventoryStatus: "source-backed",
			IntegrationLevel:    "source-backed",
			Installed:           pathExists(filepath.Join(workspaceRoot, "submodules", "tormentnexus")),
		},
		metadataHarness("opencode", "OpenCode CLI harness", "Beta", "External CLI"),
		metadataHarness("antigravity", "Antigravity CLI harness", "Experimental", "Desktop IDE / command surface"),
		metadataHarness("claude", "Claude CLI harness", "Experimental", "External CLI"),
		metadataHarness("aider", "Aider harness", "Beta", "External CLI"),
		metadataHarness("cursor", "Cursor shell harness", "Experimental", "Editor shell bridge"),
		metadataHarness("continue", "Continue CLI harness", "Experimental", "External CLI"),
		metadataHarness("cody", "Cody CLI harness", "Experimental", "External CLI"),
		metadataHarness("copilot", "GitHub Copilot CLI harness", "Experimental", "External CLI"),
		metadataHarness("adrenaline", "Adrenaline CLI harness", "Experimental", "External CLI"),
		metadataHarness("amazon-q", "Amazon Q CLI harness", "Experimental", "External CLI"),
		metadataHarness("amazon-q-developer", "Amazon Q Developer CLI harness", "Experimental", "External CLI"),
		metadataHarness("amp-code", "Amp Code CLI harness", "Experimental", "External CLI"),
		metadataHarness("auggie", "Auggie CLI harness", "Experimental", "External CLI"),
		metadataHarness("azure-openai", "Azure OpenAI CLI harness", "Experimental", "External CLI"),
		metadataHarness("bito", "Bito CLI harness", "Experimental", "External CLI"),
		metadataHarness("byterover", "Byterover CLI harness", "Experimental", "External CLI"),
		metadataHarness("claude-code", "Claude Code CLI harness", "Beta", "External CLI"),
		metadataHarness("code-codex", "Code Codex CLI harness", "Experimental", "External CLI"),
		metadataHarness("codebuff", "Codebuff harness", "Experimental", "External CLI"),
		metadataHarness("codemachine", "Codemachine harness", "Experimental", "External CLI"),
		metadataHarness("codex", "Codex CLI harness", "Beta", "External CLI"),
		metadataHarness("crush", "Crush CLI harness", "Experimental", "External CLI"),
		metadataHarness("dolt", "Dolt CLI harness", "Experimental", "External CLI"),
		metadataHarness("factory", "Factory CLI harness", "Experimental", "External CLI"),
		metadataHarness("factory-droid", "Factory Droid harness", "Experimental", "External CLI"),
		metadataHarness("gemini", "Gemini CLI harness", "Experimental", "External CLI"),
		metadataHarness("goose", "Goose harness", "Experimental", "External CLI"),
		metadataHarness("grok", "Grok CLI harness", "Experimental", "External CLI"),
		metadataHarness("jules", "Jules CLI harness", "Experimental", "External CLI"),
		metadataHarness("kilo-code", "Kilo Code CLI harness", "Experimental", "External CLI"),
		metadataHarness("kimi", "Kimi CLI harness", "Experimental", "External CLI"),
		metadataHarness("llm", "LLM CLI harness", "Experimental", "External CLI"),
		metadataHarness("litellm", "LiteLLM CLI harness", "Experimental", "External CLI"),
		metadataHarness("llamafile", "Llamafile CLI harness", "Experimental", "External CLI"),
		metadataHarness("manus", "Manus CLI harness", "Experimental", "External CLI"),
		metadataHarness("mistral-vibe", "Mistral Vibe CLI harness", "Experimental", "External CLI"),
		metadataHarness("ollama", "Ollama CLI harness", "Experimental", "External CLI"),
		metadataHarness("open-interpreter", "Open Interpreter CLI harness", "Experimental", "External CLI"),
		metadataHarness("pi", "Pi CLI harness", "Experimental", "External CLI"),
		metadataHarness("qwen-code", "Qwen Code CLI harness", "Experimental", "External CLI"),
		metadataHarness("rowboatx", "RowboatX CLI harness", "Experimental", "External CLI"),
		metadataHarness("rovo", "Rovo CLI harness", "Experimental", "External CLI"),
		metadataHarness("shell-pilot", "Shell Pilot CLI harness", "Experimental", "External CLI"),
		metadataHarness("smithery", "Smithery CLI harness", "Experimental", "External CLI"),
		metadataHarness("superai-cli", "SuperAI CLI harness", "Experimental", "External CLI"),
		metadataHarness("trae", "Trae CLI harness", "Experimental", "External CLI"),
		metadataHarness("warp", "Warp CLI harness", "Experimental", "External CLI"),
		{
			ID:                  "custom",
			Description:         "Operator-supplied custom harness",
			Maturity:            "Experimental",
			Runtime:             "Operator-defined",
			ParityNotes:         "Operator-defined harness; tool calls are not enumerable unless the operator supplies a bridge contract.",
			ToolInventoryStatus: "operator-defined",
			IntegrationLevel:    "operator-defined",
			Installed:           false,
		},
	}

	return definitions
}

func Summarize(definitions []Definition) Summary {
	summary := Summary{}
	for _, definition := range definitions {
		switch definition.ToolInventoryStatus {
		case "source-backed":
			summary.SourceBackedHarnessCount++
			summary.SourceBackedToolCount += definition.ToolCallCount
		case "operator-defined":
			summary.OperatorDefinedHarnessCount++
		default:
			summary.MetadataOnlyHarnessCount++
		}
	}
	return summary
}

func pathExists(target string) bool {
	_, err := os.Stat(target)
	return err == nil
}

func tormentnexusToolNames(workspaceRoot string) []string {
	toolsDir := filepath.Join(workspaceRoot, "submodules", "tormentnexus", "tools")
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		return nil
	}

	namePattern := regexp.MustCompile(`Name:\s*"([^"]+)"`)
	names := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(toolsDir, entry.Name()))
		if err != nil {
			continue
		}

		for _, match := range namePattern.FindAllStringSubmatch(string(content), -1) {
			if len(match) < 2 || match[1] == "" {
				continue
			}
			names[match[1]] = struct{}{}
		}
	}

	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}
