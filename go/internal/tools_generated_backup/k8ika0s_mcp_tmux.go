package tools

import (
	"context"
	"os/exec"
	"strings"
)

func HandleSendKeys(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	keys, _ :=getString(args, "keys")
	if target == "" {
		return err("target is required")
}

	cmd := exec.CommandContext(ctx, "tmux", "send-keys", "-t", target, keys)
	out, e := cmd.Output()
	if e != nil {
		return err(e.Error())
}

	return ok(strings.TrimSpace(string(out)))
}

func HandleListSessions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "tmux", "list-sessions")
	out, e := cmd.Output()
	if e != nil {
		return err(e.Error())
}

	return ok(strings.TrimSpace(string(out)))
}