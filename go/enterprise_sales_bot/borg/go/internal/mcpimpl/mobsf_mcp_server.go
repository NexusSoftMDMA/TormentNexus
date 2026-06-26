package mcpimpl

import (
    "context"
    "io"
    "net/http"
)

func HandleMobsfStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    serverURL, _ :=getString(args, "server_url")
    if serverURL == "" {
        return err("server_url is required")
}

    resp, e := http.DefaultClient.Get(serverURL + "/api/v1/health")
    if e != nil {
        return err("request failed: " + e.Error())
}

    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    return ok(string(body))
}

func HandleMobsfVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    serverURL, _ :=getString(args, "server_url")
    if serverURL == "" {
        return err("server_url is required")
}

    resp, e := http.DefaultClient.Get(serverURL + "/api/v1/version")
    if e != nil {
        return err("request failed: " + e.Error())
}

    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    return ok(string(body))
}