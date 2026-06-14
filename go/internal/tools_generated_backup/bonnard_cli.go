package tools

import (
	"bytes"
	"context"
	"net/http"
	"os/exec"
)

func HandleRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	e := cmd.Run()
	if e != nil {
		return err("execution failed: " + e.Error())
}

	return ok(out.String())
}