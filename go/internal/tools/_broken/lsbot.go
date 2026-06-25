package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func HandleConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	switch action {
	case "get":
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return err(homeErr.Error())
}

		configPath := homeDir + "/.lsbot.yaml"
		data, readErr := os.ReadFile(configPath)
		if readErr != nil {
			return err("Config file not found at " + configPath)
}

		return ok(string(data))
}
	case "set":
		key, _ :=getString(args, "key")
		val, _ :=getString(args, "value")
		if key == "" {
			return err("key is required")
}

		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return err(homeErr.Error())
}

		configPath := homeDir + "/.lsbot.yaml"
		data, readErr := os.ReadFile(configPath)
		lines := []string{}
		if readErr == nil {
			lines = strings.Split(string(data), "\n")

		re := regexp.MustCompile(`^` + regexp.QuoteMeta(key) + `:\s*.*$`)
		found := false
		for i, line := range lines {
			if re.MatchString(line) {
				lines[i] = key + ": " + val
				found = true
				break
			}
		}
		if !found {
			lines = append(lines, key+": "+val)

		content := strings.Join(lines, "\n")
		writeErr := os.WriteFile(configPath, []byte(content), 0644)
		if writeErr != nil {
			return err(writeErr.Error())
}

		return ok("Config key '" + key + "' set to '" + val + "'")
}
	case "list":
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return err(homeErr.Error())
}

		configPath := homeDir + "/.lsbot.yaml"
		data, readErr := os.ReadFile(configPath)
		if readErr != nil {
			return ok("No config file found at " + configPath)
}

		return ok(string(data))
}
	default:
		return err("Unknown action: " + action + ". Use 'get', 'set', or 'list'")

}

func HandleOnboard(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	provider, _ :=getString(args, "provider")
	apiKey, _ :=getString(args, "api_key")
	model, _ :=getString(args, "model")
	baseURL, _ :=getString(args, "base_url")

	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return err(homeErr.Error())
}

	configPath := homeDir + "/.lsbot.yaml"

	providerBlock := ""
	if provider != "" {
		providerBlock = "  provider: " + provider + "\n"
	}
	apiKeyBlock := ""
	if apiKey != "" {
		apiKeyBlock = "  api_key: " + apiKey + "\n"
	}
	modelBlock := ""
	if model != "" {
		modelBlock = "  model: " + model + "\n"
	}
	baseURLBlock := ""
	if baseURL != "" {
		baseURLBlock = "  base_url: " + baseURL + "\n"
	}

	configContent := "ai:\n" + providerBlock + apiKeyBlock + modelBlock + baseURLBlock
	writeErr := os.WriteFile(configPath, []byte(configContent), 0644)
	if writeErr != nil {
		return err(writeErr.Error())
}

	return ok("Onboarding complete. Config written to " + configPath + ".\nYou can now run lsbot with your configured provider.")
}

func HandleProviders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	type providerInfo struct {
		Name       string `json:"name"`
		Default    string `json:"default_model"`
		BaseURL    string `json:"base_url"`
		KeyURL     string `json:"api_key_url"`
		Aliases    string `json:"aliases"`
	}

	providers := []providerInfo{
		{"deepseek", "deepseek-chat", "https://api.deepseek.com/v1", "https://platform.deepseek.com/api_keys", ""},
		{"qwen", "qwen-plus", "https://dashscope.aliyuncs.com/compatible-mode/v1", "https://bailian.console.aliyun.com/", "qianwen, tongyi"},
		{"claude", "claude-sonnet-4-20250514", "https://api.anthropic.com/v1", "https://console.anthropic.com/", "anthropic"},
		{"kimi", "kimi-k2.5", "https://api.moonshot.cn/v1", "https://platform.moonshot.cn/", "moonshot"},
		{"minimax", "MiniMax-Text-01", "https://api.minimax.chat/v1", "https://platform.minimaxi.com/", ""},
		{"doubao", "doubao-pro-32k", "https://ark.cn-beijing.volces.com/api/v3", "https://console.volcengine.com/ark", "bytedance, volcengine"},
		{"zhipu", "glm-4-flash", "https://open.bigmodel.cn/api/paas/v4", "https://open.bigmodel.cn/", "glm, chatglm"},
		{"openai", "gpt-4o", "https://api.openai.com/v1", "https://platform.openai.com/api-keys", "gpt, chatgpt"},
		{"gemini", "gemini-2.0-flash", "https://generativelanguage.googleapis.com/v1beta", "https://aistudio.google.com/apikey", "google"},
		{"yi", "yi-large", "https://api.lingyiwanwu.com/v1", "https://platform.lingyiwanwu.com/", "lingyiwanwu, wanwu"},
		{"stepfun", "step-2-16k", "https://api.stepfun.com/v1", "https://platform.stepfun.com/", ""},
		{"baichuan", "Baichuan4", "https://api.baichuan-ai.com/v1", "https://platform.baichuan-ai.com/", ""},
		{"spark", "generalv3.5", "https://spark-api-open.xf-yun.com/v1", "https://console.xfyun.cn/", "iflytek, xunfei"},
		{"siliconflow", "Qwen/Qwen2.5-72B-Instruct", "https://api.siliconflow.cn/v1", "https://cloud.siliconflow.cn/", ""},
		{"grok", "grok-2-latest", "https://api.x.ai/v1", "https://console.x.ai/", "xai"},
		{"ollama", "llama3.2", "http://localhost:11434/v1", "No API key needed", ""},
	}

	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Name < providers[j].Name
	})

	output := "Supported AI Providers (16 total):\n\n"
	for i, p := range providers {
		output += fmt.Sprintf("%d. **%s**\n", i+1, p.Name)
		output += fmt.Sprintf("   Default Model: %s\n", p.Default)
		output += fmt.Sprintf("   Base URL: %s\n", p.BaseURL)
		output += fmt.Sprintf("   API Key: %s\n", p.KeyURL)
		if p.Aliases != "" {
			output += fmt.Sprintf("   Aliases: %s\n", p.Aliases)

		output += "\n"
	}

	return ok(output)
}

}

func HandleGh(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	argsList := []string{}

	subcommand, _ :=getString(args, "subcommand")
	repo, _ :=getString(args, "repo")

	if repo != "" {
		argsList = append(argsList, "--repo", repo)

	switch subcommand {
	case "pr_list":
		argsList = append(argsList, "pr", "list")
		state, _ :=getString(args, "state")
		if state != "" {
			argsList = append(argsList, "--state", state)

		limit, _ :=getInt(args, "limit")
		if limit > 0 {
			argsList = append(argsList, "--limit", strconv.Itoa(limit))

		argsList = append(argsList, "--json", "number,title,state,user")
	case "pr_view":
		number, _ :=getInt(args, "number")
		if number == 0 {
			return err("pr_number is required for pr_view")
}

		argsList = append(argsList, "pr", "view", strconv.Itoa(number))
	case "pr_checks":
		number, _ :=getInt(args, "number")
		if number == 0 {
			return err("pr_number is required for pr_checks")
}

		argsList = append(argsList, "pr", "checks", strconv.Itoa(number))
	case "run_list":
		argsList = append(argsList, "run", "list")
		limit, _ :=getInt(args, "limit")
		if limit > 0 {
			argsList = append(argsList, "--limit", strconv.Itoa(limit))

	case "run_view":
		runID, _ :=getString(args, "run_id")
		if runID == "" {
			return err("run_id is required for run_view")
}

		argsList = append(argsList, "run", "view", runID)
		logFailed, _ :=getBool(args, "log_failed")
		if logFailed {
			argsList = append(argsList, "--log-failed")

	case "issue_list":
		argsList = append(argsList, "issue", "list")
		state, _ :=getString(args, "state")
		if state != "" {
			argsList = append(argsList, "--state", state)

		argsList = append(argsList, "--json", "number,title")
	case "api":
		endpoint, _ :=getString(args, "endpoint")
		if endpoint == "" {
			return err("endpoint is required for api subcommand")
}

		argsList = append(argsList, "api", endpoint)
		jq, _ :=getString(args, "jq")
		if jq != "" {
			argsList = append(argsList, "--jq", jq)

	default:
		return err("Unknown subcommand: " + subcommand)
}

	cmd := exec.CommandContext(ctx, "gh", argsList...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err("gh error: " + string(output) + " - " + execErr.Error())
}

	return ok(string(output))
}

}
}
}
}
}
}
}

func HandleMcporter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	serverName, _ :=getString(args, "server")

	argsList := []string{}

	switch action {
	case "list":
		if serverName != "" {
			argsList = append(argsList, "list", serverName)
			schema, _ :=getBool(args, "schema")
			if schema {
				argsList = append(argsList, "--schema")

		} else {
			argsList = append(argsList, "list")

	case "call":
		toolName, _ :=getString(args, "tool")
		if toolName == "" {
			return err("tool is required for call action")
}

		argsList = append(argsList, "call", toolName)
		toolArgs, _ :=getString(args, "args")
		if toolArgs != "" {
			argsList = append(argsList, toolArgs)

		toolJSON, _ :=getString(args, "args_json")
		if toolJSON != "" {
			argsList = append(argsList, "--args", toolJSON)

	case "auth":
		target, _ :=getString(args, "target")
		if target == "" {
			return err("target is required for auth action")
}

		argsList = append(argsList, "auth", target)
		reset, _ :=getBool(args, "reset")
		if reset {
			argsList = append(argsList, "--reset")

	case "config":
		configAction, _ :=getString(args, "config_action")
		switch configAction {
		case "list":
			argsList = append(argsList, "config", "list")
		case "get":
			key, _ :=getString(args, "key")
			if key == "" {
				return err("key is required for config get")
}

			argsList = append(argsList, "config", "get", key)
		case "add":
			key, _ :=getString(args, "key")
			val, _ :=getString(args, "value")
			if key == "" || val == "" {
				return err("key and value are required for config add")
}

			argsList = append(argsList, "config", "add", key, val)
		case "remove":
			key, _ :=getString(args, "key")
			if key == "" {
				return err("key is required for config remove")
}

			argsList = append(argsList, "config", "remove", key)
		default:
			return err("Unknown config action: " + configAction)
}

	case "daemon":
		daemonAction, _ :=getString(args, "daemon_action")
		argsList = append(argsList, "daemon", daemonAction)
	default:
		return err("Unknown action: " + action)
}

	cmd := exec.CommandContext(ctx, "mcporter", argsList...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err("mcporter error: " + string(output) + " - " + execErr.Error())
}

	return ok(string(output))
}

}
}
}
}
}

func HandleReviewCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	branch, _ :=getString(args, "branch")
	if branch == "" {
		branch = "main"
	}

	// Get diff of current branch compared to base branch
	cmd := exec.CommandContext(ctx, "git", "diff", branch+"...HEAD", "--stat")
	statOutput, statErr := cmd.CombinedOutput()
	if statErr != nil {
		return err("git diff stat error: " + statErr.Error())
}

	// Get full diff
	cmd2 := exec.CommandContext(ctx, "git", "diff", branch+"...HEAD")
	diffOutput, diffErr := cmd2.CombinedOutput()
	if diffErr != nil {
		return err("git diff error: " + diffErr.Error())
}

	// Get changed files list
	cmd3 := exec.CommandContext(ctx, "git", "diff", branch+"...HEAD", "--name-only")
	filesOutput, filesErr := cmd3.CombinedOutput()
	if filesErr != nil {
		return err("git diff name-only error: " + filesErr.Error())
}

	// Get commit log
	cmd4 := exec.CommandContext(ctx, "git", "log", branch+"..HEAD", "--oneline")
	logOutput, logErr := cmd4.CombinedOutput()
	if logErr != nil {
		return err("git log error: " + logErr.Error())
}

	changedFiles := strings.TrimSpace(string(filesOutput))
	commits := strings.TrimSpace(string(logOutput))
	stat := strings.TrimSpace(string(statOutput))
	diff := strings.TrimSpace(string(diffOutput))

	if diff == "" {
		return ok("No changes found on current branch compared to " + branch + ".")
}

	// Parse changed files for analysis
	fileList := strings.Split(changedFiles, "\n")

	report := fmt.Sprintf("# Code Review: Current branch vs `%s`\n\n", branch)
	report += "## Commits\n" + commits + "\n\n"
	report += "## Changed Files\n" + changedFiles + "\n\n"
	report += "## Diff Stats\n```\n" + stat + "\n```\n\n"

	// Security checks
	securityIssues := []string{}
	securityPatterns := []string{
		`(?i)(api[_-]?key|secret|password|token)\s*[:=]\s*['"][^'"]+['"]`,")
		`(?i)eval\s*\(`,
		`(?i)exec\s*\(`,
		`(?i)os\.system\s*\(`,
		`(?i)SELECT\s+.*\+.*FROM`,
		`(?i)INSERT\s+.*\+.*INTO`,
	}
	for _, pattern := range securityPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(diff) {
			securityIssues = append(securityIssues, "Pattern match: "+pattern)

	}

	report += "## Review Summary\n\n"
	report += fmt.Sprintf("- **Files changed**: %d\n", len(fileList))
	report += fmt.Sprintf("- **Commits ahead**: %d\n", strings.Count(commits, "\n")+1)

	if len(securityIssues) > 0 {
		report += "\n### ⚠️ Security Concerns\n"
		for _, issue := range securityIssues {
			report += "- " + issue + "\n"
		}
	} else {
		report += "\n### ✅ No obvious security issues detected\n"
	}

	report += "\n### 📝 Notes\n"
	report += "- Review the diff manually for correctness and design quality\n"
	report += "- Ensure tests are included for new functionality\n"
	report += "- Check for proper error handling in new code\n"

	if len(diff) > 4000 {
		report += "\n### Diff (truncated)\n```\n" + diff[:4000] + "\n...\n```\n"
	} else {
		report += "\n### Diff\n```\n" + diff + "\n```\n"
	}

	return ok(report)
}

}

func HandlePeekaboo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	if action == "" {
		return err("action is required for peekaboo")
}

	peekArgs := []string{}

	switch action {
	case "permissions":
		peekArgs = append(peekArgs, "permissions")
	case "list":
		listType, _ :=getString(args, "type")
		if listType == "" {
			return err("type is required for list (apps, windows, screens, menubar, permissions)")
}

		peekArgs = append(peekArgs, "list", listType)
	case "see":
		peekArgs = append(peekArgs, "see")
		appName, _ :=getString(args, "app")
		if appName != "" {
			peekArgs = append(peekArgs, "--app", appName)

		annotate, _ :=getBool(args, "annotate")
		if annotate {
			peekArgs = append(peekArgs, "--annotate")

		path, _ :=getString(args, "path")
		if path != "" {
			peekArgs = append(peekArgs, "--path", path)

	case "click":
		peekArgs = append(peekArgs, "click")
		target, _ :=getString(args, "target")
		if target != "" {
			peekArgs = append(peekArgs, "--on", target)

		appName, _ :=getString(args, "app")
		if appName != "" {
			peekArgs = append(peekArgs, "--app", appName)

	case "type":
		text, _ :=getString(args, "text")
		if text == "" {
			return err("text is required for type action")
}

		peekArgs = append(peekArgs, "type", text)
		appName, _ :=getString(args, "app")
		if appName != "" {
			peekArgs = append(peekArgs, "--app", appName)

		retKey, _ :=getBool(args, "return")
		if retKey {
			peekArgs = append(peekArgs, "--return")

	case "app":
		appAction, _ :=getString(args, "app_action")
		if appAction == "" {
			return err("app_action is required for app (launch, quit, relaunch, hide, unhide, switch)")
}

		peekArgs = append(peekArgs, "app", appAction)
		appName, _ :=getString(args, "app")
		if appName != "" {
			peekArgs = append(peekArgs, "--app", appName)

	case "image":
		peekArgs = append(peekArgs, "image")
		path, _ :=getString(args, "path")
		if path != "" {
			peekArgs = append(peekArgs, "--path", path)

	default:
		return err("Unknown peekaboo action: " + action)
}

	cmd := exec.CommandContext(ctx, "peekaboo", peekArgs...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err("peekaboo error: " + string(output) + " - " + execErr.Error())
}

	return ok(string(output))
}

}
}
}
}
}
}
}
}
}

func HandleGateway(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	provider, _ :=getString(args, "provider")
	apiKey, _ :=getString(args, "api_key")
	model, _ :=getString(args, "model")
	baseURL, _ :=getString(args, "base_url")
	platform, _ :=getString(args, "platform")
	port, _ :=getInt(args, "port")

	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return err(homeErr.Error())
}

	configPath := homeDir + "/.lsbot.yaml"

	// Read existing config
	existingData, readErr := os.ReadFile(configPath)
	existing := string(existingData)
	if readErr != nil {
		existing = ""
	}

	// Build config sections
	aiBlock := "ai:\n"
	if provider != "" {
		aiBlock += "  provider: " + provider + "\n"
	}
	if apiKey != "" {
		aiBlock += "  api_key: " + apiKey + "\n"
	}
	if model != "" {
		aiBlock += "  model: " + model + "\n"
	}
	if baseURL != "" {
		aiBlock += "  base_url: " + baseURL + "\n"
	}

	relayBlock := ""
	if platform != "" {
		relayBlock = "relay:\n  platform: " + platform + "\n"
		if provider != "" {
			relayBlock += "  provider: " + provider + "\n"
		}
	}

	configContent := aiBlock + "\n" + relayBlock

	// Preserve existing overrides if present
	if strings.Contains(existing, "overrides:") {
		overrideIdx := strings.Index(existing, "overrides:")
		configContent += "\n" + existing[overrideIdx:]
	}

	writeErr := os.WriteFile(configPath, []byte(configContent), 0644)
	if writeErr != nil {
		return err(writeErr.Error())
}

	portStr := "8080"
	if port > 0 {
		portStr = strconv.Itoa(port)

	// Start gateway
	cmd := exec.CommandContext(ctx, "lsbot", "gateway", "--port", portStr)
	cmd.Env = append(os.Environ())
	startErr := cmd.Start()
	if startErr != nil {
		return err("Failed to start gateway: " + startErr.Error())
}

	return ok(fmt.Sprintf("lsbot gateway started on port %s (PID: %d) with provider: %s, platform: %s", portStr, cmd.Process.Pid, provider, platform))
}
}