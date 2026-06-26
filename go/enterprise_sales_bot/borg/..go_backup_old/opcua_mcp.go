package tools

import "context"

func HandleReadNode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    _ = ctx
    nodeId, _ :=getString(args, "nodeId")
    if nodeId == "" {
        return err("nodeId is required")
}

    return ok("Read node: " + nodeId)
}