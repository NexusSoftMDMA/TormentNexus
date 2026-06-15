package tools

import (
	"context"
	"encoding/json"
)

// HandleGetAgents returns a list of agents.
func HandleGetAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agents := []map[string]interface{}{
		{"id": "47", "name": "Agent 47", "status": "active"},
		{"id": "48", "name": "Agent 48", "status": "retired"},
	}
	b, e := json.Marshal(agents)
	if e != nil {
		return err("failed to marshal agents")
	}
	return success(string(b))
}

// HandleGetAgent returns details for a single agent.
func HandleGetAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
	}
	agent := map[string]string{"id": id, "name": "Agent " + id, "status": "active"}
	b, e := json.Marshal(agent)
	if e != nil {
		return err("failed to marshal agent")
	}
	return success(string(b))
}