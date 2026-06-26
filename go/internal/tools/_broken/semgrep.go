package tools

import (
    "context"
    "os/exec"
    "strings"
)

// We assume that parity.go defines:
// 
// func ok(text string) (ToolResponse, error)
// func err(msg string) (ToolResponse, error)
// func getString(args map[string]interface{}, key string) string
// etc.

func HandleSemgrepVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // Run: semgrep --version
    cmd := exec.CommandContext(ctx, "semgrep", "--version")
    output, apiErr := cmd.CombinedOutput()
    if apiErr != nil {
        return err(string(output))
}

    return ok(strings.TrimSpace(string(output)))
}

func HandleSemgrepScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // Get the path to scan
    path, _ :=getString(args, "path")
    if path == "" {
        return err("missing required argument: path")
}

    // Optional: config, rules, etc.
    // For simplicity, just run: semgrep scan <path>
    cmd := exec.CommandContext(ctx, "semgrep", "scan", path)
    output, apiErr := cmd.CombinedOutput()
    if apiErr != nil {
        return err(string(output))
}

    return ok(string(output))
}