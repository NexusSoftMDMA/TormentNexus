package tools

import (
    "context"
    "io"
    "net/http"
)

func HandleS2Search(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query is required")
}

    resp, e := http.DefaultClient.Get("https://api.example.com/search?q=" + query)
    if e != nil {
        return err("API call failed: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("read failed: " + e.Error())
}

    return success(string(body))
}

func HandleS2GetDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    id, _ :=getString(args, "id")
    if id == "" {
        return err("id is required")
}

    resp, e := http.DefaultClient.Get("https://api.example.com/item/" + id)
    if e != nil {
        return err("API call failed: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("read failed: " + e.Error())
}

    return success(string(body))
}