package mcpimpl

import "context"

func HandleListSignals(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "all"
	}
	return ok("Listed signals: " + name)
}

func HandleGetSignal_root_signals_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "signal_id")
	if id == "" {
		return err("signal_id is required")
}

	return success("Signal " + id + " retrieved")
}