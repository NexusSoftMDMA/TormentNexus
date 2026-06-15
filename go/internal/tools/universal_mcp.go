//go:build ignore
// +build ignore

package tools

import (
    "context"
    "io/ioutil"
    "net/http"
)

func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url is required")
}

    resp, e := http.DefaultClient.Get(url)
    if e != nil {
        return err("fetch failed: " + e.Error())
}

    defer resp.Body.Close()
    body, e := ioutil.ReadAll(resp.Body)
    if e != nil {
        return err("read failed: " + e.Error())
}

    return success(string(body))
}
