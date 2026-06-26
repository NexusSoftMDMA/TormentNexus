package tools

import (
	"context"
	"encoding/json"
	"strings"
)

func HandleAnalyzeCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("missing code argument")
}

	nodes := strings.Fields(code)
	graph := map[string][]string{"root": nodes}
	data, e := json.Marshal(graph)
	if e != nil {
		return err("failed to marshal graph")
}

	return success(string(data))
}

func HandleCompleteCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("missing prompt argument")
}

	suggestion := "completion for " + prompt
	return success(suggestion)
}// touch 1781132122
