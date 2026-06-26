package tools

import (
    "context"
    "fmt"
    "os/exec"
    "time"
)

func HandleExecutePython(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    code, _ :=getString(args, "code")
    if code == "" {
        return err("code is required")
}

    timeout, _ :=getInt(args, "timeout")
    if timeout <= 0 {
        timeout = 30
    }
    ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
    defer cancel()
    cmd := exec.CommandContext(ctxTimeout, "python3", "-c", code)
    output, e := cmd.CombinedOutput()
    if e != nil {
        return err(fmt.Sprintf("execution error: %v: %s", e, string(output)))
}

    return success(string(output))
}