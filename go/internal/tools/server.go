package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// HandleEcho returns the message provided in the arguments.
func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ := getString(args, "message")
	return ok(message)
}

// HandleHelloWorld returns a greeting message.
func HandleHelloWorld(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ := getString(args, "name")
	if name == "" {
		name = "World"
	}

	response := map[string]string{
		"message": fmt.Sprintf("Hello, %s!", name),
	}

	data, e := json.Marshal(response)
	if e != nil {
		return err(e.Error())
	}

	return ok(string(data))
}
