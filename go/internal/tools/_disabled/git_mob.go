package tools

import (
	"context"
	"os/exec"
	"strings"
)

func HandleGetMob(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("git", "config", "--get", "mob.current")
	out, e := cmd.Output()
	if e != nil {
		return ok("No one is mobbing.")
}

	members := strings.TrimSpace(string(out))
	return ok("Current mob: " + members)
}

func HandleSetMob(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	members, _ :=getString(args, "members")
	if members == "" {
		return err("members is required")
}

	cmd := exec.Command("git", "config", "--add", "mob.current", members)
	e := cmd.Run()
	if e != nil {
		return err("Failed to set mob: " + e.Error())
}

	return ok("Mob set to " + members)
}