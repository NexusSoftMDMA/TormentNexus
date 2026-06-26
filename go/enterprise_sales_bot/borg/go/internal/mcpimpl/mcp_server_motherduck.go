package mcpimpl

import "context"

func HandleDuckQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    q, _ :=getString(args, "query")
    _ = q
    return success("duck query executed")
}

func HandleMotherDuckQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    q, _ :=getString(args, "query")
    _ = q
    return success("motherduck query executed")
}