package tools

import (
	"bytes"
	"context"
	"os/exec"
)

func HandleRunJavaScript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code argument is required")
}

	cmd := exec.CommandContext(ctx, "node", "-e", code)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	e := cmd.Run()
	if e != nil {
		return err("execution error: " + stderr.String() + ": " + e.Error())
}

	return ok(stdout.String())
}