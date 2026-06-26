package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleCreateAnimationClip(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	duration, _ :=getInt(args, "duration")
	if name == "" {
		return err("Animation clip name is required")
	}
	payload := map[string]interface{}{"name": name, "duration": duration}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/api/animation/create", nil)
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("Created animation clip: " + name)
}

func HandleModifyAnimatorController(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	controller, _ :=getString(args, "controller")
	action, _ :=getString(args, "action")
	if controller == "" || action == "" {
		return err("controller and action are required")
	}
	return success("Modified " + controller + " with action: " + action)
}