package tools

import (
	"context"
	"net/http"
)

func HandleBrickCreate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	color, _ :=getString(args, "color")
	if name == "" {
		return err("name is required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.brick.example.com/bricks", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err("unexpected status: " + resp.Status)
}

	return ok("Brick " + name + " (" + color + ") created")
}