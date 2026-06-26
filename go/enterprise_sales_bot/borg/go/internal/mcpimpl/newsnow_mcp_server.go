package mcpimpl

import (
    "context"
    "net/http"
    "encoding/json"
)

func HandleGetNews_newsnow_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    category, _ :=getString(args, "category")
    if category == "" {
        return err("category is required")
}

    data := map[string]string{"news": "Top headlines", "category": category}
    body, e := json.Marshal(data)
    if e != nil {
        return err("failed to marshal")
}

    return ok(string(body))
}

func HandleSearchNews_newsnow_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query is required")
}

    resp, e := http.DefaultClient.Get("https://api.example.com/search?q=" + query)
    if e != nil {
        return err("request failed")
}

    defer resp.Body.Close()
    return ok("search performed for " + query)
}