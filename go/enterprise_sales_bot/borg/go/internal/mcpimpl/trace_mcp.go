package mcpimpl

import "context"

func HandleTrace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	traceID, _ :=getString(args, "trace_id")
	spanID, _ :=getString(args, "span_id")
	return ok("Trace ID: " + traceID + ", Span ID: " + spanID)
}

func HandleMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverName, _ :=getString(args, "server")
	return ok("MCP server " + serverName + " is running")
}