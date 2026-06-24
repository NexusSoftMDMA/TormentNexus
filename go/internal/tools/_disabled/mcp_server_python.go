package tools

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "net/http"
)

func HandleExecutePython(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    code, _ :=getString(args, "code")
    if code == "" {
        return err("code is required")
}

    reqBody, _ := json.Marshal(map[string]string{"code": code})
    resp, e := http.DefaultClient.Post("https://python-executor.example.com/run", "application/json", bytes.NewReader(reqBody))
    if e != nil {
        return err("request error: " + e.Error())
}

    defer resp.Body.Close()
    result, _ := io.ReadAll(resp.Body)
    return ok(string(result))
}