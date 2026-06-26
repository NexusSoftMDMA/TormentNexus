package mcpimpl

import "context"

func HandleGreet_model_context_protocol_advanced_ai_agents_for_beginners(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name + "!")
}

func HandleCurrentTime_model_context_protocol_advanced_ai_agents_for_beginners(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	format, _ :=getString(args, "format")
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return ok("Current time: " + time.Now().Format(format))
}