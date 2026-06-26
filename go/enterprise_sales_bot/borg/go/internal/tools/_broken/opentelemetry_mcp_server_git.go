package mcpimpl

import "context"

func HandleListTraces_opentelemetry_mcp_server_git(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("OpenTelemetry traces list")
}

func HandleGetTrace_opentelemetry_mcp_server_git(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Trace details for ID: " + getString(args, "traceId"))
}