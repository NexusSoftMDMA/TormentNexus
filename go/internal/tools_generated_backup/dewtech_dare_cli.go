package tools

import "context"

func HandleDareVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("DARE CLI version 1.0.0")
}

func HandleDareHelp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Available DARE commands: query, graph, init, version")
}