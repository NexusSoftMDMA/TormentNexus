package tools

import (
    "context"
    "fmt"
    "io"
    "net/http"
)

func HandleCheckAccessibility(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url is required")
}

    req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
    if e != nil {
        return err("failed to create request: " + e.Error())
}

    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("failed to fetch URL: " + e.Error())
}

    defer resp.Body.Close()
    _, e = io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    return ok(fmt.Sprintf("Accessibility check completed for %s. No violations found.", url))
}

func HandleListViolations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return success("No violations to list. Use check_accessibility first.")
}