package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version := "1.0-java-sdk"
	name, _ :=getString(args, "name")
	if name == "" {
		name = "JavaMCP"
	}
	data := map[string]string{"name": name, "version": version}
	body, e := json.Marshal(data)
	if e != nil {
		return err("failed to encode version")
}

	return ok(string(body))
}

func HandleCapabilities(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	caps := []string{"tools", "resources", "prompts", "sampling"}
	body, e := json.Marshal(caps)
	if e != nil {
		return err("failed to encode capabilities")
}

	return ok(string(body))
}