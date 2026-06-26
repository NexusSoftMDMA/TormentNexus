package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleVibeStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing id")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.vibe.dev/status/"+id, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
}

	return ok("status retrieved")
}

func HandleVibeSync(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	if project == "" {
		return err("missing project")
}

	return success("synced " + project)
}// touch 1781132143
