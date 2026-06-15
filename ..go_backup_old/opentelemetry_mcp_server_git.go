package tools

import "context"

func HandleListTraces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("OpenTelemetry traces list")
}

func HandleGetTrace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Trace details for ID: " + getString(args, "traceId"))
}