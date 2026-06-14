package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandlePangoLint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query argument")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.pangolint.com/search?q="+query, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return success("PangoLint knowledge retrieved")
}

func HandlePangoScript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	script, _ :=getString(args, "script")
	if script == "" {
		return err("missing script argument")
	}
	return ok("PangoScript details for " + script)
}// touch 1781132137
