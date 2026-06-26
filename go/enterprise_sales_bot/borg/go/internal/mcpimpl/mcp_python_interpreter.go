package mcpimpl

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func HandleExecutePython_mcp_python_interpreter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("'code' argument is required")
}

	cmd := exec.CommandContext(ctx, "python3", "-c", code)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	e := cmd.Run()
	if e != nil {
		return err(fmt.Sprintf("execution error: %s\nstderr: %s", e.Error(), stderr.String()))
}

	return success(fmt.Sprintf("Output:\n%s", strings.TrimSpace(stdout.String())))
}

func HandleGetPythonVersion_mcp_python_interpreter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "python3", "--version")
	output, e := cmd.Output()
	if e != nil {
		return err(fmt.Sprintf("failed to get Python version: %s", e.Error()))
}

	return ok(fmt.Sprintf("Python version: %s", strings.TrimSpace(string(output))))
}