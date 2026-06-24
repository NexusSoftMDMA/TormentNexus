package tools

import "context"

func HandleCreatePool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	poolName, _ :=getString(args, "pool_name")
	if poolName == "" {
		return err("pool_name is required")
}

	maxAgents, _ :=getInt(args, "max_agents")
	if maxAgents <= 0 {
		return err("max_agents must be positive")
}

	return ok("pool created: " + poolName)
}

func HandleMatch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	poolID, _ :=getString(args, "pool_id")
	if poolID == "" {
		return err("pool_id is required")
}

	agentID, _ :=getString(args, "agent_id")
	if agentID == "" {
		return err("agent_id is required")
}

	return ok("agent " + agentID + " matched in pool " + poolID)
}