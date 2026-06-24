package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

func HandleListEndpoints(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url is required")
}

    resp, e := http.DefaultClient.Get(url)
    if e != nil {
        return err("failed to fetch spec: " + e.Error())
}

    defer resp.Body.Close()
    var spec map[string]interface{}
    if e := json.NewDecoder(resp.Body).Decode(&spec); e != nil {
        return err("failed to parse JSON: " + e.Error())
}

    paths, found := spec["paths"].(map[string]interface{})
    if !found {
        return err("no paths in spec")
}

    count := 0
    for range paths {
        count++
    }
    return ok(fmt.Sprintf("Found %d paths", count))
}