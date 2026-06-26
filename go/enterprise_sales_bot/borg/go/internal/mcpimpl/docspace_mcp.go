package mcpimpl

import (
    "context"
    "io"
    "net/http"
)

func HandleSearchDocuments_docspace_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query parameter required")
}

    resp, e := http.DefaultClient.Get("https://api.docspace.example.com/search?q=" + query)
    if e != nil {
        return err("search request failed: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    return success(string(body))
}