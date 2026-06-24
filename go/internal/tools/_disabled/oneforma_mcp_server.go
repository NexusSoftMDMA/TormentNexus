package tools

import (
    "context"
    "os"
    "os/exec"
)

func ExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    cmd, _ :=getString(args, "command")
    if cmd == "" {
        return err("command is required")
}

    c := exec.CommandContext(ctx, "sh", "-c", cmd)
    out, e := c.CombinedOutput()
    if e != nil {
        return err("command failed: " + string(out) + ": " + e.Error())
}

    return ok(string(out))
}

func WriteFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    path, _ :=getString(args, "path")
    content, _ :=getString(args, "content")
    if path == "" {
        return err("path is required")
}

    e := os.WriteFile(path, []byte(content), 0644)
    if e != nil {
        return err("write failed: " + e.Error())
}

    return ok("file written successfully")
}