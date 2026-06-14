package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func CreateWorkspace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.brxce.com/workspaces", nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	req.Body = http.NoBody
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("api call failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err("workspace creation failed")
}

	return success("workspace created")
}

func ListWorkspaces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.brxce.com/workspaces")
	if e != nil {
		return err(fmt.Sprintf("api call failed: %v", e))
}

	defer resp.Body.Close()
	var result []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("found %d workspaces", len(result)))
}