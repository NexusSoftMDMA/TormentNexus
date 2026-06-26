package mcpimpl

import (
    "context"
)

func HandleListNetworks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("ListNetworks handler invoked")
}

func HandleGetNetwork(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    networkCode, _ :=getString(args, "networkCode")
    if networkCode == "" {
        return err("networkCode is required")
}

    return success("GetNetwork handler for code: " + networkCode)
}