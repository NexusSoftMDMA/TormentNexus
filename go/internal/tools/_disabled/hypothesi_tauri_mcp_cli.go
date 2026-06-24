package tools

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

func HandleRunTauri(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	argsStr, _ :=getString(args, "args")
	var cmdArgs []string
	if argsStr != "" {
		cmdArgs = strings.Fields(argsStr)

	c := exec.Command("tauri", append([]string{cmd}, cmdArgs...)...)
	var outBuf, errBuf bytes.Buffer
	c.Stdout = &outBuf
	c.Stderr = &errBuf
	e := c.Run()
	if e != nil {
		return err("failed to execute tauri command: " + e.Error() + "\nstderr: " + errBuf.String())
}

	return ok(outBuf.String())
}
}