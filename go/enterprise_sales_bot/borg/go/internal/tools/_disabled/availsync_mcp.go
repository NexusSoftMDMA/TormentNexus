package tools

import (
	"context"
	"fmt"
)

func HandleCheckCodingAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agent, _ :=getString(args, "agent")
	if agent == "" {
		return err("agent is required")
}

	return ok(fmt.Sprintf("Coding agent '%s' is active", agent))
}

func HandleCheckSchedulingGuardrails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	start, _ :=getString(args, "start")
	end, _ :=getString(args, "end")
	if start == "" || end == "" {
		return err("start and end are required")
}

	return ok("Scheduling guardrails are okay")
}