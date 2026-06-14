package tools

import (
    "context"
    "io"
    "net/http"
)

func HandleSearchScreenpipe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query is required")
}

    url := "http://localhost:3030/search?q=" + query
    req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
    if e != nil {
        return err(e.Error())
}

    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err(e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err(e.Error())
}

    return ok(string(body))
}