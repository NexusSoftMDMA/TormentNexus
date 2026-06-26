package mcpimpl

import (
	"context"
)

func HandleGetEspInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deviceID, _ :=getString(args, "device_id")
	if deviceID == "" {
		return err("device_id is required")
	}
	return success("ESP device " + deviceID + " info: firmware v2.1, status online")
}

func HandleSendEspCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deviceID, _ :=getString(args, "device_id")
	command, _ :=getString(args, "command")
	if deviceID == "" || command == "" {
		return err("device_id and command are required")
	}
	return success("Command '" + command + "' sent to " + deviceID + " successfully")
}