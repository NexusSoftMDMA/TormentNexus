package tools

import (
    "context"
    "io"
    "net/http"
)

func HandleSave(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url is required")
    }
    resp, e := http.DefaultClient.Get("https://web.archive.org/save/" + url)
    if e != nil {
        return err("save failed: " + e.Error())
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return err("unexpected status: " + resp.Status)
    }
    return ok("page saved")
}

func HandleGetSnapshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url is required")
    }
    resp, e := http.DefaultClient.Get("https://archive.org/wayback/available?url=" + url)
    if e != nil {
        return err("request failed: " + e.Error())
    }
    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("read failed: " + e.Error())
    }
    return success(string(body))
}// touch 1781132134
