package tools

import (
    "context"
    "encoding/json"
    "io"
    "net/http"
)

func HandleAdfinQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query is required")
}

    resp, e := http.DefaultClient.Get("https://api.adfin.example.com/search?q=" + query)
    if e != nil {
        return err("request failed: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("read failed: " + e.Error())
}

    var result map[string]interface{}
    if e := json.Unmarshal(body, &result); e != nil {
        return err("invalid JSON: " + e.Error())
}

    return ok(result["response"].(string))
}