package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleListAgents_cz_agents_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agents := []string{"agent1", "agent2", "agent3"}
	data, e := json.Marshal(agents)
	if e != nil {
		return err("failed to encode agents")
}

	return success(string(data))
}