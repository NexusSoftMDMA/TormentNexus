package tools

import (
	"context"
	"os/exec"
	"strings"
)

func HandleOpenApplication(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("application name required")
}

	e := exec.Command("cmd", "/c", "start", name).Run()
	if e != nil {
		return err("failed to open: " + e.Error())
}

	return ok("application started")
}

func HandleListProcesses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	out, e := exec.Command("tasklist").Output()
	if e != nil {
		return err("failed to list processes: " + e.Error())
}

	return success(strings.TrimSpace(string(out)))
}