package mcpimpl

import (
    "context"
    "io/ioutil"
    "net/http"
)

func HandleListMCP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url required")
}

    resp, e := http.DefaultClient.Get(url)
    if e != nil {
        return err("get failed: " + e.Error())
}

    defer resp.Body.Close()
    body, e := ioutil.ReadAll(resp.Body)
    if e != nil {
        return err("read failed: " + e.Error())
}

    return ok(string(body))
}

func HandleGetMCPInfo_mcp_router(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url required")
}

    resp, e := http.DefaultClient.Get(url)
    if e != nil {
        return err("get failed: " + e.Error())
}

    defer resp.Body.Close()
    body, e := ioutil.ReadAll(resp.Body)
    if e != nil {
        return err("read failed: " + e.Error())
}

    return ok(string(body))
}