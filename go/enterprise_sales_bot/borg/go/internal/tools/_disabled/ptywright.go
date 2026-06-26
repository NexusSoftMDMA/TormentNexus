package tools

import (
	"context"
	"os/exec"
)

func HandlePtyRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdStr, _ :=getString(args, "command")
	if cmdStr == "" {
		return err("command is required")
}

	c := exec.Command("/bin/sh", "-c", cmdStr)
	out, e := c.CombinedOutput()
	if e != nil {
		return err("execution failed: " + e.Error())
}

	return ok("output: " + string(out))
}

func HandlePtyWrite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("input is required")
}

	return ok("echo: " + input)
}