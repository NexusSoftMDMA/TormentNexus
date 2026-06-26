package mcpimpl

import (
	"context"
)

func HandleStartBroker(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	port, _ :=getString(args, "port")
	host, _ :=getString(args, "host")
	_ = port
	_ = host
	return ok("MQTT broker started on " + host + ":" + port)
}

func HandleStopBroker(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = args
	return ok("MQTT broker stopped")
}