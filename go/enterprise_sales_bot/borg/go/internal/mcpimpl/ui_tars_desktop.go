package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleGetWindows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	windows := []string{"Window1", "Window2", "Window3"}
	data, e := json.Marshal(windows)
	if e != nil {
		return err("failed to marshal windows")
}

	return success(string(data))
}

func HandleClick_ui_tars_desktop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title is required")
}

	msg := fmt.Sprintf("Clicked on window: %s", title)
	return ok(msg)
}