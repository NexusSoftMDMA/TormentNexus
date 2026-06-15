//go:build ignore
// +build ignore

package tools

import "context"

func HandleListAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name != "" {
        return ok("Agent: " + name)
}

    return ok("All agents")
}
