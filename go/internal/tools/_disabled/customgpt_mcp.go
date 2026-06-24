package tools

import "context"

func HandleCustomgpt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok("Custom GPT processed: " + query)
}

func HandleCustomgptList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Available custom GPTs: demo_gpt_1, demo_gpt_2")
}