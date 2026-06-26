package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleQuerySessions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	return success(fmt.Sprintf("Perfetto command '%s' submitted", cmd))
}

func HandleListTraces_perfetto_mcp_rs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	traces := []string{"trace1", "trace2"}
	data, _ := json.Marshal(traces)
	return ok(string(data))
}