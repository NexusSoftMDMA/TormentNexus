//go:build ignore
// +build ignore

package tools

import (
	"context"
)

func HandleListAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ :=getString(args, "network")
	if network == "" {
		return err("network is required")
}

	return success("agents in " + network + ": agent1, agent2")
}

func HandleGetAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "agent_id")
	if id == "" {
		return err("agent_id is required")
}

	return ok("agent " + id + " is online")
}
