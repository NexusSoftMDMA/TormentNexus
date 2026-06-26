package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Hello from Consult7!")
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return ok(fmt.Sprintf("Echo: %s", msg))
}

func HandleJson(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("input is required")
}

	var data map[string]interface{}
	if e := json.Unmarshal([]byte(input), &data); e != nil {
		return err(fmt.Sprintf("invalid JSON: %v", e))
}

	return ok(fmt.Sprintf("Parsed JSON: %+v", data))
}