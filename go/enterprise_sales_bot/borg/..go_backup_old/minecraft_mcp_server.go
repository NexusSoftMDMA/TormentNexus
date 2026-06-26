package tools

import "context"

func HandleSendCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	return success("Command sent: " + cmd)
}

func HandleGetPosition(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(`{"x": 100, "y": 64, "z": 200}`)
}