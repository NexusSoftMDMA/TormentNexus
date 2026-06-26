package tools

import (
	"context"
	"os/exec"
)

func HandleRunMegaLinter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	config, _ :=getString(args, "config")
	flavor, _ :=getString(args, "flavor")
	env, _ :=getString(args, "env")

	cmdArgs := []string{
		"npx", "mega-linter-runner",
		"--path", path,
	}
	if config != "" {
		cmdArgs = append(cmdArgs, "--config", config)

	if flavor != "" {
		cmdArgs = append(cmdArgs, "--flavor", flavor)

	if env != "" {
		cmdArgs = append(cmdArgs, "--env", env)

	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("megalinter failed: " + e.Error())
	}
	return ok(string(out))
}
}
}
}