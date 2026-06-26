package controlplane

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Tool struct {
	Type         string   `json:"type"`
	Name         string   `json:"name"`
	Command      string   `json:"command"`
	Available    bool     `json:"available"`
	Version      string   `json:"version,omitempty"`
	Path         string   `json:"path,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

type ToolProvider interface {
	DetectAll(context.Context) ([]Tool, error)
}

type definition struct {
	Type         string
	Name         string
	Command      string
	VersionArgs  []string
	Capabilities []string
}

type Detector struct {
	definitions []definition
	timeout     time.Duration
	ttl         time.Duration

	mu       sync.Mutex
	inflight chan struct{}
	cached   []Tool
	cachedAt time.Time
}

func NewDetector(timeout, ttl time.Duration) *Detector {
	return &Detector{
		timeout: timeout,
		ttl:     ttl,
		definitions: []definition{
			{Type: "go", Name: "Go", Command: "go", VersionArgs: []string{"version"}, Capabilities: []string{"build", "test", "server"}},
			{Type: "node", Name: "Node.js", Command: "node", VersionArgs: []string{"--version"}, Capabilities: []string{"runtime", "scripts"}},
			{Type: "python", Name: "Python", Command: "python", VersionArgs: []string{"--version"}, Capabilities: []string{"runtime", "scripts"}},
			{Type: "tormentnexus", Name: "tormentnexus CLI", Command: "tormentnexus", VersionArgs: []string{"version"}, Capabilities: []string{"chat", "edit", "repl", "tormentnexus-adapter"}},
			{Type: "antigravity", Name: "Antigravity CLI", Command: "antigravity", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "desktop", "editor", "automation"}},
			{Type: "opencode", Name: "OpenCode CLI", Command: "opencode", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "edit", "multi-file", "autonomous"}},
			{Type: "claude", Name: "Claude CLI", Command: "claude", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "analyze"}},
			{Type: "aider", Name: "Aider CLI", Command: "aider", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "edit", "git-aware", "multi-file"}},
			{Type: "cursor", Name: "Cursor CLI", Command: "cursor", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "editor"}},
			{Type: "continue", Name: "Continue CLI", Command: "continue", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "autocomplete", "editor"}},
			{Type: "cody", Name: "Cody CLI", Command: "cody", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "search", "code"}},
			{Type: "copilot", Name: "GitHub Copilot CLI", Command: "github-copilot-cli", VersionArgs: []string{"--version"}, Capabilities: []string{"explain", "suggest", "chat", "terminal", "shell"}},
			{Type: "adrenaline", Name: "Adrenaline CLI", Command: "adrenaline", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "shell", "automation"}},
			{Type: "amazon-q", Name: "Amazon Q CLI", Command: "q", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "aws", "deploy"}},
			{Type: "amazon-q-developer", Name: "Amazon Q Developer CLI", Command: "q-developer", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "aws"}},
			{Type: "amp-code", Name: "Amp Code CLI", Command: "amp", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "edit", "terminal"}},
			{Type: "auggie", Name: "Auggie CLI", Command: "auggie", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "review", "git"}},
			{Type: "azure-openai", Name: "Azure OpenAI CLI", Command: "az-openai", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "azure"}},
			{Type: "bito", Name: "Bito CLI", Command: "bito", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "review"}},
			{Type: "byterover", Name: "Byterover CLI", Command: "byterover", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "automation"}},
			{Type: "claude-code", Name: "Claude Code CLI", Command: "claude-code", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "analyze"}},
			{Type: "code-codex", Name: "Code CLI (Codex fork)", Command: "code-codex", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "edit"}},
			{Type: "codex", Name: "Codex CLI", Command: "codex", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code"}},
			{Type: "codebuff", Name: "Codebuff CLI", Command: "codebuff", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "refactor"}},
			{Type: "codemachine", Name: "Codemachine CLI", Command: "codemachine", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "generate"}},
			{Type: "crush", Name: "Crush CLI", Command: "crush", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "data"}},
			{Type: "dolt", Name: "Dolt CLI", Command: "dolt", VersionArgs: []string{"version"}, Capabilities: []string{"database", "sql", "git"}},
			{Type: "factory", Name: "Factory CLI", Command: "factory", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "build"}},
			{Type: "factory-droid", Name: "Factory Droid CLI", Command: "factory-droid", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "build"}},
			{Type: "gemini", Name: "Gemini CLI", Command: "gemini", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "multimodal"}},
			{Type: "goose", Name: "Goose CLI", Command: "goose", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "agent"}},
			{Type: "grok", Name: "Grok CLI", Command: "grok", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "realtime"}},
			{Type: "jules", Name: "Jules CLI", Command: "jules", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "agent"}},
			{Type: "kilo-code", Name: "Kilo Code CLI", Command: "kilo", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "editor"}},
			{Type: "kimi", Name: "Kimi CLI", Command: "kimi", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "long-context"}},
			{Type: "llm", Name: "LLM CLI", Command: "llm", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "models", "prompt"}},
			{Type: "litellm", Name: "LiteLLM CLI", Command: "litellm", VersionArgs: []string{"--version"}, Capabilities: []string{"models", "proxy", "routing"}},
			{Type: "llamafile", Name: "Llamafile CLI", Command: "llamafile", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "local", "models"}},
			{Type: "manus", Name: "Manus CLI", Command: "manus", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "agent", "automation"}},
			{Type: "mistral-vibe", Name: "Mistral Vibe CLI", Command: "mistral", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "local"}},
			{Type: "ollama", Name: "Ollama CLI", Command: "ollama", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "local", "models"}},
			{Type: "open-interpreter", Name: "Open Interpreter CLI", Command: "interpreter", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "python", "shell"}},
			{Type: "pi", Name: "Pi CLI", Command: "pi", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "personal", "assistant"}},
			{Type: "qwen-code", Name: "Qwen Code CLI", Command: "qwen", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "local"}},
			{Type: "rowboatx", Name: "RowboatX CLI", Command: "rowboatx", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "data"}},
			{Type: "rovo", Name: "Rovo CLI", Command: "rovo", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "search"}},
			{Type: "shell-pilot", Name: "Shell Pilot CLI", Command: "shell-pilot", VersionArgs: []string{"--version"}, Capabilities: []string{"shell", "chat", "automation"}},
			{Type: "smithery", Name: "Smithery CLI", Command: "smithery", VersionArgs: []string{"--version"}, Capabilities: []string{"mcp", "registry", "tools"}},
			{Type: "superai-cli", Name: "SuperAI CLI", Command: "superai", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "automation"}},
			{Type: "trae", Name: "Trae CLI", Command: "trae", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "code", "review"}},
			{Type: "warp", Name: "Warp CLI", Command: "warp", VersionArgs: []string{"--version"}, Capabilities: []string{"chat", "terminal", "collaborative"}},
		},
	}
}

func (d *Detector) DetectAll(ctx context.Context) ([]Tool, error) {
	d.mu.Lock()
	if len(d.cached) > 0 && time.Since(d.cachedAt) < d.ttl {
		tools := append([]Tool(nil), d.cached...)
		d.mu.Unlock()
		return tools, nil
	}

	if d.inflight != nil {
		wait := d.inflight
		d.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-wait:
			d.mu.Lock()
			tools := append([]Tool(nil), d.cached...)
			d.mu.Unlock()
			return tools, nil
		}
	}

	wait := make(chan struct{})
	d.inflight = wait
	d.mu.Unlock()

	tools := make([]Tool, len(d.definitions))
	var wg sync.WaitGroup
	for i, def := range d.definitions {
		wg.Add(1)
		go func(idx int, d2 definition) {
			defer wg.Done()
			tools[idx] = d.detectTool(ctx, d2)
		}(i, def)
	}
	wg.Wait()

	d.mu.Lock()
	d.cached = append([]Tool(nil), tools...)
	d.cachedAt = time.Now()
	close(wait)
	d.inflight = nil
	d.mu.Unlock()

	return tools, nil
}

func (d *Detector) detectTool(ctx context.Context, def definition) Tool {
	tool := Tool{
		Type:         def.Type,
		Name:         def.Name,
		Command:      def.Command,
		Available:    false,
		Capabilities: append([]string(nil), def.Capabilities...),
	}

	// Check if it's a relative path or a command in PATH
	executable, err := exec.LookPath(def.Command)
	if err != nil {
		// Fallback: check if the path exists in the current directory
		if _, err := os.Stat(def.Command); err == nil {
			executable = def.Command
		} else {
			return tool
		}
	}

	tool.Available = true
	tool.Path = executable

	commandCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	output, err := exec.CommandContext(commandCtx, executable, def.VersionArgs...).CombinedOutput()
	if err != nil && commandCtx.Err() != nil {
		tool.Version = "timeout"
		return tool
	}

	tool.Version = parseVersion(string(output))
	return tool
}

var versionPattern = regexp.MustCompile(`v?(\d+\.\d+(?:\.\d+)?)`)

func parseVersion(output string) string {
	match := versionPattern.FindStringSubmatch(output)
	if len(match) >= 2 {
		return match[1]
	}

	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return "unknown"
	}

	return trimmed
}
