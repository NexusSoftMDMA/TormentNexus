package tools

import (
	"context"
	"fmt"
)

func HandleCreateObservation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	content, _ :=getString(args, "content")
	if name == "" {
		return err("name is required")
}

	msg := fmt.Sprintf("Observation '%s' created with content: %s", name, content)
	_, found := args["content"]
	if !found {
		content = "no content"
	}
	return success(msg + " (content found: " + content + ")")
}

func HandleGetObservation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	return ok("Observation retrieved for id: " + id)
}