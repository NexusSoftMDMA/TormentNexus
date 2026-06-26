package mcpimpl

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
)

func HandleCreateTunnel_sharnix_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    port, _ :=getInt(args, "port")
    url := fmt.Sprintf("https://api.sharnix.com/tunnels?name=%s&port=%d", name, port)
    req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
    if e != nil {
        return err("failed to create request")
    }
    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("failed to create tunnel: " + e.Error())
    }
    defer resp.Body.Close()
    body, e := ioutil.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response")
    }
    if resp.StatusCode != 200 {
        return err("API error: " + string(body))
    }
    return success(string(body))
}

func HandleShareLink(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    urlValue, _ :=getString(args, "url")
    duration, _ :=getInt(args, "duration")
    apiURL := fmt.Sprintf("https://api.sharnix.com/share?url=%s&duration=%d", urlValue, duration)
    req, e := http.NewRequestWithContext(ctx, "POST", apiURL, nil)
    if e != nil {
        return err("failed to create request")
    }
    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("failed to share link: " + e.Error())
    }
    defer resp.Body.Close()
    body, e := ioutil.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response")
    }
    if resp.StatusCode != 200 {
        return err("API error: " + string(body))
    }
    return success(string(body))
}