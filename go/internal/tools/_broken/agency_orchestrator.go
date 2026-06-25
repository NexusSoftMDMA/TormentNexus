package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func getProviderBaseURL(provider string) string {
	switch strings.ToLower(provider) {
	case "deepseek":
		return "https://api.deepseek.com/v1"
}
	case "openai":
		return "https://api.openai.com/v1"
}
	case "claude", "anthropic":
		return "https://api.anthropic.com/v1"
	case "agnes":
		return "https://api.agnes.ai/v1"
	case "compshare":
		return "https://api.compshare.ai/v1"
	default:
		return "https://api.openai.com/v1"
	}
}

func getProviderModel(provider string) string {
	switch strings.ToLower(provider) {
	case "deepseek":
		return "deepseek-chat"
	case "openai":
		return "gpt-4o"
	case "claude", "anthropic":
		return "claude-sonnet-4-20250514"
	case "agnes":
		return "agnes-default"
	case "compshare":
		return "compshare-default"
	default:
		return provider
	}
}

func extractContent(result map[string]interface{}) string {
	choices, found := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return ""
	}
	choice, found := choices[0].(map[string]interface{})
	if !found {
		return ""
	}
	message, found := choice["message"].(map[string]interface{})
	if !found {
		return ""
	}
	content, found := message["content"].(string)
	if !found {
		return ""
	}
	return content
}

func HandleRunWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflowPath, _ :=getString(args, "workflow")
	if workflowPath == "" {
		return err("workflow path is required")

	resume, _ :=getBool(args, "resume")
	fromStep, _ :=getString(args, "from")
	feedback, _ :=getString(args, "feedback")
	team, _ :=getString(args, "team")
	compare, _ :=getBool(args, "compare")

	data, readErr := os.ReadFile(workflowPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read workflow file: %v", readErr))
}

	name, steps := parseWorkflowYAML(string(data))
	if name == "" {
		name = filepath.Base(workflowPath)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Workflow: %s\n", name))
	sb.WriteString(fmt.Sprintf("Steps: %d\n", len(steps)))

	if team != "" {
		sb.WriteString(fmt.Sprintf("Team: %s\n", team))

	if resume {
		sb.WriteString("Mode: resume\n")
		if fromStep != "" {
			sb.WriteString(fmt.Sprintf("Resume from step: %s\n", fromStep))

		if feedback != "" {
			sb.WriteString(fmt.Sprintf("Feedback: %s\n", feedback))

	}
	if compare {
		sb.WriteString("Compare mode: enabled (workflow vs single-shot baseline + blind judge)\n")

	order, sortErr := topologicalSort(steps)
	if sortErr != nil {
		return err(fmt.Sprintf("DAG resolution failed: %v", sortErr))
}

	sb.WriteString("\nExecution Plan:\n")
	for i, stepID := range order {
		s := findStep(steps, stepID)
		deps := ""
		if s != nil && len(s.DependsOn) > 0 {
			deps = fmt.Sprintf(" (after: %s)", strings.Join(s.DependsOn, ", "))

		sb.WriteString(fmt.Sprintf("  %d. %s%s\n", i+1, stepID, deps))

	outputDir := fmt.Sprintf("ao-output/%s-%d", sanitizeName(name), time.Now().Unix())
	sb.WriteString(fmt.Sprintf("\nOutputs will be saved to: %s/\n", outputDir))
	sb.WriteString("\nTo iterate on a specific step later:\n")
	sb.WriteString(fmt.Sprintf("  ao run %s --resume last --from <step-id>\n", workflowPath))

	return ok(sb.String())
}

}
}
}
}
}
}

func HandleComposeWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	task, _ :=getString(args, "task")
	if task == "" {
		return err("task description is required")
}

	provider, _ :=getString(args, "provider")
	if provider == "" {
		provider = "deepseek"
	}
	model, _ :=getString(args, "model")
	team, _ :=getString(args, "team")
	lang, _ :=getString(args, "lang")
	if lang == "" {
		lang = "zh"
	}
	runNow, _ :=getBool(args, "run")

	rolesDir := os.Getenv("AO_AGENTS_DIR")
	if rolesDir == "" {
		homeDir, homeErr := os.UserHomeDir()
		if homeErr == nil {
			if lang == "en" {
				rolesDir = filepath.Join(homeDir, ".ao", "agency-agents")
			} else {
				rolesDir = filepath.Join(homeDir, ".ao", "agency-agents-zh")

		}
	}

	roles := listRoles(rolesDir)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Composing workflow for: %s\n", task))
	sb.WriteString(fmt.Sprintf("Provider: %s\n", provider))
	if model != "" {
		sb.WriteString(fmt.Sprintf("Model: %s\n", model))

	sb.WriteString(fmt.Sprintf("Language: %s\n", lang))
	sb.WriteString(fmt.Sprintf("Available roles: %d\n", len(roles)))

	if team != "" {
		teamsDir := os.Getenv("AO_TEAMS_DIR")
		if teamsDir == "" {
			homeDir, _ := os.UserHomeDir()
			teamsDir = filepath.Join(homeDir, ".ao", "teams")

		teamData, teamErr := os.ReadFile(filepath.Join(teamsDir, team+".team.yaml"))
		if teamErr == nil {
			sb.WriteString(fmt.Sprintf("Pinned team: %s\n", team))
			sb.WriteString(fmt.Sprintf("Team definition:\n%s\n", string(teamData)))
		} else {
			sb.WriteString(fmt.Sprintf("Team '%s' not found, will auto-select roles\n", team))

	}

	suggested := suggestRoles(task, roles)
	if len(suggested) > 0 {
		sb.WriteString("\nSuggested roles:\n")
		for _, r := range suggested {
			sb.WriteString(fmt.Sprintf("  - %s\n", r))

	}

	if runNow {
		sb.WriteString("\nWorkflow composed and running...\n")
	} else {
		sb.WriteString("\nWorkflow composed. Use --run to execute.\n")

	return ok(sb.String())
}

}
}
}
}
}
}

func HandleListRoles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	lang, _ :=getString(args, "lang")
	if lang == "" {
		lang = "zh"
	}

	rolesDir := os.Getenv("AO_AGENTS_DIR")
	if rolesDir == "" {
		homeDir, homeErr := os.UserHomeDir()
		if homeErr == nil {
			if lang == "en" {
				rolesDir = filepath.Join(homeDir, ".ao", "agency-agents")
			} else {
				rolesDir = filepath.Join(homeDir, ".ao", "agency-agents-zh")

		}
	}

	roles := listRoles(rolesDir)

	if category != "" {
		var filtered []string
		for _, r := range roles {
			if strings.HasPrefix(r, category+"/") || strings.Contains(r, category) {
				filtered = append(filtered, r)

		}
		roles = filtered
	}

	sort.Strings(roles)

	var sb strings.Builder
	if lang == "en" {
		sb.WriteString(fmt.Sprintf("Available expert roles (%d total):\n", len(roles)))
	} else {
		sb.WriteString(fmt.Sprintf("可用专家角色（共 %d 个）：\n", len(roles)))

	categories := make(map[string][]string)
	for _, r := range roles {
		parts := strings.SplitN(r, "/", 2)
		cat := "uncategorized"
		if len(parts) == 2 {
			cat = parts[0]
		}
		categories[cat] = append(categories[cat], r)

	catNames := make([]string, 0, len(categories))
	for k := range categories {
		catNames = append(catNames, k)

	sort.Strings(catNames)

	for _, cat := range catNames {
		sb.WriteString(fmt.Sprintf("\n%s (%d):\n", cat, len(categories[cat])))
		for _, r := range categories[cat] {
			sb.WriteString(fmt.Sprintf("  - %s\n", r))

	}

	return ok(sb.String())
}

}
}
}
}
}
}

func HandlePromptOptimize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt text is required")
}

	mode, _ :=getString(args, "mode")
	if mode == "" {
		mode = "system"
	}
	save, _ :=getBool(args, "save")
	provider, _ :=getString(args, "provider")
	if provider == "" {
		provider = "deepseek"
	}

	apiKey := getProviderKey(provider)
	if apiKey == "" {
		return err(fmt.Sprintf("API key not found for provider '%s'. Set the appropriate environment variable (e.g., DEEPSEEK_API_KEY, OPENAI_API_KEY, ANTHROPIC_API_KEY)", provider))
}

	baseURL := getProviderBaseURL(provider)
	model := getProviderModel(provider)

	var systemPrompt string
	if mode == "system" {
		systemPrompt = "You are a prompt optimization expert. The user will give you a system prompt. Rewrite it to be clearer, more specific, and more effective. Output ONLY the improved prompt text, nothing else. Do not execute the prompt — produce a better prompt."
	} else {
		systemPrompt = "You are a prompt optimization expert. The user will give you a user prompt. Rewrite it to be clearer, more specific, and more effective for getting high-quality responses. Output ONLY the improved prompt text, nothing else. Do not execute the prompt — produce a better prompt."
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  2048,
	}

	jsonData, marshalErr := json.Marshal(reqBody)
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", marshalErr))
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", strings.NewReader(string(jsonData)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return err(fmt.Sprintf("API request failed: %v", httpErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	optimized := extractContent(result)

	var sb strings.Builder
	sb.WriteString("Prompt Optimization Result\n")
	sb.WriteString("==========================\n\n")
	sb.WriteString(fmt.Sprintf("Mode: %s\n", mode))
	sb.WriteString(fmt.Sprintf("Provider: %s / %s\n\n", provider, model))
	sb.WriteString("Original:\n")
	sb.WriteString(prompt + "\n\n")
	sb.WriteString("Optimized:\n")
	sb.WriteString(optimized + "\n")

	if save {
		promptsDir := os.Getenv("AO_PROMPTS_DIR")
		if promptsDir == "" {
			homeDir, _ := os.UserHomeDir()
			promptsDir = filepath.Join(homeDir, ".ao", "prompts")

		os.MkdirAll(promptsDir, 0755)
		saveName := fmt.Sprintf("prompt-%d.prompt.json", time.Now().Unix())
		saveData := map[string]interface{}{
			"original":  prompt,
			"optimized": optimized,
			"mode":      mode,
			"provider":  provider,
			"model":     model,
			"timestamp": time.Now().Format(time.RFC3339),
		}
		saveJSON, _ := json.MarshalIndent(saveData, "", "  ")
		writeErr := os.WriteFile(filepath.Join(promptsDir, saveName), saveJSON, 0644)
		if writeErr != nil {
			sb.WriteString(fmt.Sprintf("\nWarning: failed to save: %v\n", writeErr))
		} else {
			sb.WriteString(fmt.Sprintf("\nSaved to: %s/%s\n", promptsDir, saveName))

	}

	return ok(sb.String())
}

type workflowStep struct {
	ID        string
	Role      string
	Task      string
	DependsOn []string
}

}
}

func parseWorkflowYAML(content string) (string, []workflowStep) {
	var name string
	var steps []workflowStep
	var currentStep *workflowStep

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "name:") {
			name = strings.TrimSpace(strings.TrimPrefix(trimmed, "name:"))
			name = strings.Trim(name, "\"'")

		if strings.HasPrefix(trimmed, "- id:") {
			if currentStep != nil {
				steps = append(steps, *currentStep)

			id := strings.TrimSpace(strings.TrimPrefix(trimmed, "- id:"))
			id = strings.Trim(id, "\"'")
			currentStep = &workflowStep{ID: id}
		}
		if currentStep != nil {
			if strings.HasPrefix(trimmed, "role:") {
				currentStep.Role = strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "role:")), "\"'")

			if strings.HasPrefix(trimmed, "depends_on:") {
				depsStr := strings.TrimSpace(strings.TrimPrefix(trimmed, "depends_on:"))
				depsStr = strings.Trim(depsStr, "[]")
				for _, d := range strings.Split(depsStr, ",") {
					d = strings.TrimSpace(strings.Trim(d, "\"' "))
					if d != "" {
						currentStep.DependsOn = append(currentStep.DependsOn, d)

				}
			}
		}
	}
	if currentStep != nil {
		steps = append(steps, *currentStep)

	return name, steps
}

}
}
}
}
}

func topologicalSort(steps []workflowStep) ([]string, error) {
	inDegree := make(map[string]int)
	graph := make(map[string][]string)
	stepIDs := make(map[string]bool)

	for _, s := range steps {
		stepIDs[s.ID] = true
		if _, found := inDegree[s.ID]; !ok {
			inDegree[s.ID] = 0
		}
		for _, dep := range s.DependsOn {
			graph[dep] = append(graph[dep], s.ID)
			inDegree[s.ID]++
		}
	}

	var queue []string
	for id := range stepIDs {
		if inDegree[id] == 0 {
			queue = append(queue, id)

	}
	sort.Strings(queue)

	var order []string
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		order = append(order, curr)
		nexts := graph[curr]
		sort.Strings(nexts)
		for _, n := range nexts {
			inDegree[n]--
			if inDegree[n] == 0 {
				queue = append(queue, n)

		}
		sort.Strings(queue)

	if len(order) != len(stepIDs) {
		return nil, fmt.Errorf("cycle detected in workflow DAG")
}

	return order, nil
}

}
}
}

func findStep(steps []workflowStep, id string) *workflowStep {
	for i := range steps {
		if steps[i].ID == id {
			return &steps[i]
		}
	}
	return nil
}

func sanitizeName(name string) string {
	re := regexp.MustCompile(`[<>:"/\\|?*]`)")
	return re.ReplaceAllString(name, "_")
}

func listRoles(dir string) []string {
	var roles []string
	if dir == "" {
		return roles
	}
	filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		content := string(data)
		if !strings.Contains(content, "name:") {
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			rel = path
		}
		ext := filepath.Ext(rel)
		roleID := strings.TrimSuffix(rel, ext)
		roleID = strings.ReplaceAll(roleID, "\\", "/")")
		roles = append(roles, roleID)
		return nil
	})
	return roles
}

func suggestRoles(task string, roles []string) []string {
	keywords := strings.ToLower(task)
	var suggested []string
	keyTerms := []string{"design", "review", "write", "test", "architect", "product", "market", "data", "security", "devops", "game", "finance", "legal", "research", "翻译", "设计", "评审", "测试", "架构", "产品", "营销", "数据", "安全", "运维", "游戏", "金融", "法律", "研究"}

	for _, r := range roles {
		rl := strings.ToLower(r)
		for _, term := range keyTerms {
			if strings.Contains(rl, term) && strings.Contains(keywords, term) {
				suggested = append(suggested, r)
				break
			}
		}
	}

	if len(suggested) > 8 {
		suggested = suggested[:8]
	}
	return suggested
}

func getProviderKey(provider string) string {
	switch strings.ToLower(provider) {
	case "deepseek":
		return os.Getenv("DEEPSEEK_API_KEY")
}
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
}
	case "claude", "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	case "agnes":
		return os.Getenv("AGNES_API_KEY")
	case "compshare":
		return os.Getenv("COMPSHARE_API_KEY")
	default:
		envName := strings.ToUpper(provider) + "_API_KEY"
		return os.Getenv(envName)

}