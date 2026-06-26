package tools

import (
    "context"
    "io"
    "net/http"
    "net/url"
)

func HandleRunPagespeed(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    urlStr, _ :=getString(args, "url")
    key, _ :=getString(args, "key")
    if urlStr == "" {
        return err("url is required")
}

    if key == "" {
        return err("key is required")
}

    apiURL := "https://www.googleapis.com/pagespeedonline/v5/runPagespeed?url=" + url.QueryEscape(urlStr) + "&key=" + key
    resp, e := http.DefaultClient.Get(apiURL)
    if e != nil {
        return err("failed to call PageSpeed API: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    return success(string(body))
}