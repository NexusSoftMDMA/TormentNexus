package tools

import (
	"context"
	"os/exec"
)

func HandleListSnapshots(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repository")
	if repo == "" {
		repo = "."
	}
	cmd := exec.CommandContext(ctx, "duplicacy", "list", "-r", repo)
	out, e := cmd.Output()
	if e != nil {
		return err("failed to list snapshots: " + e.Error())
}

	return ok(string(out))
}

func HandleBackup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repository")
	if repo == "" {
		repo = "."
	}
	tag, _ :=getString(args, "tag")
	cmdArgs := []string{"backup", "-r", repo}
	if tag != "" {
		cmdArgs = append(cmdArgs, "-tag", tag)

	cmd := exec.CommandContext(ctx, "duplicacy", cmdArgs...)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("backup failed: " + e.Error() + "\n" + string(out))
}

	return ok(string(out))
}
}