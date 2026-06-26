package mcpimpl

import (
	"context"
	"fmt"
)

func HandleListPorts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ports := []string{"/dev/ttyUSB0", "/dev/ttyUSB1", "COM1", "COM2"}
	return ok(fmt.Sprintf("Available ports: %v", ports))
}

func HandleOpenSerial(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	port, _ :=getString(args, "port")
	if port == "" {
		return err("port parameter is required")
}

	return success(fmt.Sprintf("Serial port %s opened successfully (simulated)", port))
}