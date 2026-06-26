package tools

import (
	"context"
	"os/exec"
)

func HandleRunMCP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	var cmdArgs []string
	if rawArgs, found := args["args"]; found {
		if arr, found := rawArgs.([]interface{}); found {
			for _, a := range arr {
				cmdArgs = append(cmdArgs, a.(string))

		}
	}
	output, e := exec.CommandContext(ctx, cmd, cmdArgs...).Output()
	if e != nil {
		return err("failed to run command: " + e.Error())
}

	return ok(string(output))
}
}