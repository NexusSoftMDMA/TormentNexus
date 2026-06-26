package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleStepOver(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    threadId, _ :=getString(args, "threadId")
    resp, e := http.DefaultClient.Get("http://localhost:8080/debug/step?threadId=" + threadId)
    if e != nil {
        return err("stepover http request failed: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("stepover read body failed: " + e.Error())
}

    if resp.StatusCode != 200 {
        return err("stepover returned status " + fmt.Sprint(resp.StatusCode) + ": " + string(body))
}

    var result map[string]interface{}
    if e := json.Unmarshal(body, &result); e != nil {
        return err("stepover decode failed: " + e.Error())
}

    return ok("stepover completed")
}

func HandleContinue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    threadId, _ :=getString(args, "threadId")
    resp, e := http.DefaultClient.Get("http://localhost:8080/debug/continue?threadId=" + threadId)
    if e != nil {
        return err("continue http request failed: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("continue read body failed: " + e.Error())
}

    if resp.StatusCode != 200 {
        return err("continue returned status " + fmt.Sprint(resp.StatusCode) + ": " + string(body))
}

    var result map[string]interface{}
    if e := json.Unmarshal(body, &result); e != nil {
        return err("continue decode failed: " + e.Error())
}

    return ok("continue completed")
}