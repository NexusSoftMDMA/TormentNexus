package mcpimpl

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleListProcesses_winremote_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("tasklist", "/FO", "CSV", "/NH")
	out, e := cmd.Output()
	if e != nil {
		return err("failed to list processes: " + e.Error())
}

	return ok(string(out))
}

func HandleKillProcess_winremote_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pid, _ :=getInt(args, "pid")
	cmd := exec.Command("taskkill", "/PID", fmt.Sprintf("%d", pid))
	_, e := cmd.Output()
	if e != nil {
		return err("failed to kill process: " + e.Error())
}

	return success("process killed")
}