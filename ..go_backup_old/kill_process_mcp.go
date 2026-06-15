package tools

import (
	"context"
	"os"
	"syscall"
)

func HandleKillProcess(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pid, _ :=getInt(args, "pid")
	if pid <= 0 {
		return err("pid must be a positive integer")
}

	p, e := os.FindProcess(pid)
	if e != nil {
		return err("failed to find process: " + e.Error())
}

	e = p.Signal(syscall.SIGTERM)
	if e != nil {
		return err("failed to kill process: " + e.Error())
}

	return success("process killed successfully")
}