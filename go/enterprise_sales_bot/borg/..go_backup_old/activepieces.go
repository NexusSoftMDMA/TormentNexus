package tools

import (
    "context"
)

func HandleActivepieces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    action, _ :=getString(args, "action")
    if action == "info" {
        return ok("Activepieces: AI Agents & MCPs & AI Workflow Automation (~400 MCP servers)")
}

    return ok("Activepieces tool ready")
}