package tools

import (
    "context"
    "io"
    "net/http"
)

func HandleFetchUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url is required")
}

    resp, e := http.DefaultClient.Get(url)
    if e != nil {
        return err("failed to fetch: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read body: " + e.Error())
}

    return ok(string(body))
}