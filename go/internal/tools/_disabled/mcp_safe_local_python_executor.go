package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

func HandleExecutePython(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	cmd := exec.CommandContext(ctx, "python3", "-c", code)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	e := cmd.Run()
	if e != nil {
		return err(fmt.Sprintf("execution error: %s", stderr.String()))
}

	return success(stdout.String())
}