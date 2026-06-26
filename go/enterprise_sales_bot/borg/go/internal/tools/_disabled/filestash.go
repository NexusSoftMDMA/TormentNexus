package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
)

func HandleListFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    path, _ :=getString(args, "path")
    baseURL := os.Getenv("FILESTASH_URL")
    if baseURL == "" {
        return err("FILESTASH_URL not set")
}

    url := fmt.Sprintf("%s/api/v1/ls?path=%s", baseURL, path)
    req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
    if e != nil {
        return err("create request failed: " + e.Error())
}

    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("request failed: " + e.Error())
}

    defer resp.Body.Close()
    body, e := ioutil.ReadAll(resp.Body)
    if e != nil {
        return err("read failed: " + e.Error())
}

    var result interface{}
    if e := json.Unmarshal(body, &result); e != nil {
        return err("unmarshal failed: " + e.Error())
}

    return ok(fmt.Sprintf("files: %v", result))
}