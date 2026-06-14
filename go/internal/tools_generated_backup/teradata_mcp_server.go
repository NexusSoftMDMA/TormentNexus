package tools

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func executeQuery(ctx context.Context, url, query string) (string, error) {
    body, _ := json.Marshal(map[string]string{"query": query})
    req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    if e != nil {
        return "", fmt.Errorf("request creation failed: %w", e)
}

    req.Header.Set("Content-Type", "application/json")
    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return "", fmt.Errorf("request failed: %w", e)
}

    defer resp.Body.Close()
    data, e := io.ReadAll(resp.Body)
    if e != nil {
        return "", fmt.Errorf("read failed: %w", e)
}

    if resp.StatusCode != 200 {
        return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(data))
}

    return string(data), nil
}

func HandleExecuteQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    query, _ :=getString(args, "query")
    if url == "" || query == "" {
        return err("url and query are required")
}

    result, e := executeQuery(ctx, url, query)
    if e != nil {
        return err(e.Error())
}

    return ok(result)
}

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url is required")
}

    query := "SELECT table_name FROM information_schema.tables"
    result, e := executeQuery(ctx, url, query)
    if e != nil {
        return err(e.Error())
}

    return ok(result)
}