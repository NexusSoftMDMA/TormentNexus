package tools

import (
	"context"
	"os"
)

func HandleTerminateProcess(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pid, _ :=getInt(args, "pid")
	proc, e := os.FindProcess(pid)
	if e != nil {
		return err("process not found")
}

	e = proc.Kill()
	if e != nil {
		return err("kill failed: " + e.Error())
}

	return success("process terminated")
}