package mcpimpl

import (
    "context"
    "io"
    "net/http"
)

func HandleListProjects_ghidramcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url parameter required")
}

    resp, e := http.DefaultClient.Get(url + "/projects")
    if e != nil {
        return err("failed to fetch projects: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    return ok(string(body))
}

func HandleAnalyzeFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    file, _ :=getString(args, "file")
    if url == "" || file == "" {
        return err("url and file parameters required")
}

    resp, e := http.DefaultClient.Get(url + "/analyze?file=" + file)
    if e != nil {
        return err("failed to analyze: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    return ok(string(body))
}