package tools

import (
	"context"
	"fmt"
)

func HandleListDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Devices: [{'id':1,'name':'device1'}]")
}

func HandleGetDevice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	return ok(fmt.Sprintf("Device ID: %d", id))
}