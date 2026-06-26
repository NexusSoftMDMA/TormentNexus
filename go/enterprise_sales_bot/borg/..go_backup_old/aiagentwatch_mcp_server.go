package tools

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
)

func HandleSendProgress(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url required")
}

    data, _ := json.Marshal(map[string]interface{}{
        "agent_id": getString(args, "agent_id"),
        "message":  getString(args, "message"),
        "status":   getString(args, "status"),
    })
    req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
    if e != nil {
        return err("request: " + e.Error())
}

    req.Header.Set("Content-Type", "application/json")
    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("do: " + e.Error())
}

    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return err("status: " + resp.Status)
}

    return ok("progress sent")
}