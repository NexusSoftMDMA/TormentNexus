package mcpimpl

import "context"

func HandleListSystemPrompts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompts := []string{"code_review", "summarize", "explain", "debug"}
	return ok(prompts)
}

func HandleGetSystemPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	prompts := map[string]string{
		"code_review": "You are an expert code reviewer...",
		"summarize":   "Summarize the following text...",
		"explain":     "Explain the concept...",
		"debug":       "Help debug the following code...",
	}
	content, found := prompts[name]
	if !found {
		return err("unknown prompt: " + name)
}

	return success(content)
}