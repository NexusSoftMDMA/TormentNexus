package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

func HandleListCommands_gk_cli(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmds := []string{"greet", "sum", "version"}
	data, e := json.Marshal(cmds)
	if e != nil {
		return err("failed to marshal commands")
}

	return ok(string(data))
}

func HandleRunCommand_gk_cli(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdName, _ :=getString(args, "command")
	if cmdName == "" {
		return err("command is required")
}

	argsSlice := []string{}
	if argsRaw, found := args["args"]; found {
		if arr, found := argsRaw.([]interface{}); found {
			for _, a := range arr {
				if s, found := a.(string); found {
					argsSlice = append(argsSlice, s)

			}
		}
	}
	out, e := exec.CommandContext(ctx, cmdName, argsSlice...).CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("execution failed: %v", e))
}

	return ok(string(out))
}
}