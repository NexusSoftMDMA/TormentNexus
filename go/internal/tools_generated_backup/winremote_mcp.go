package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleListProcesses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("tasklist", "/FO", "CSV", "/NH")
	out, e := cmd.Output()
	if e != nil {
		return err("failed to list processes: " + e.Error())
}

	return ok(string(out))
}

func HandleKillProcess(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pid, _ :=getInt(args, "pid")
	cmd := exec.Command("taskkill", "/PID", fmt.Sprintf("%d", pid))
	_, e := cmd.Output()
	if e != nil {
		return err("failed to kill process: " + e.Error())
}

	return success("process killed")
}