package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	project, _ :=getString(args, "project")
	scheme, _ :=getString(args, "scheme")
	destination, _ :=getString(args, "destination")

	xcArgs := []string{}
	if command != "" {
		xcArgs = append(xcArgs, command)

	if project != "" {
		xcArgs = append(xcArgs, "-project", project)

	if scheme != "" {
		xcArgs = append(xcArgs, "-scheme", scheme)

	if destination != "" {
		xcArgs = append(xcArgs, "-destination", destination)

	cmd := exec.CommandContext(ctx, "xcodebuild", xcArgs...)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("xcodebuild failed: %v\n%s", e, string(out)))
}

	return success(strings.TrimSpace(string(out)))
}
}
}
}
}