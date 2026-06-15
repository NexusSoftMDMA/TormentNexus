package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearchTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := "https://genart.dev/api/tools?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch tools: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Tools []string `json:"tools"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d tools: %v", len(result.Tools), result.Tools))
}

func HandleGetPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tool, _ :=getString(args, "tool")
	url := "https://genart.dev/api/prompts?tool=" + tool
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch prompt: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Prompt string `json:"prompt"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	return ok(result.Prompt)
}