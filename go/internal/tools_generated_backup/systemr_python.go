package tools

import (
    "context"
    "os/exec"
    "strings"
)

func HandleExecutePython(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    code, _ :=getString(args, "code")
    if code == "" {
        return err("code is required")
}

    cmd := exec.Command("python3", "-c", code)
    out, e := cmd.CombinedOutput()
    if e != nil {
        return err("execution error: " + string(out) + " " + e.Error())
}

    return ok(string(out))
}