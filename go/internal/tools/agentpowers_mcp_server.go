//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "q")
	resp, e := http.DefaultClient.Get("https://api.agentpowers.com/skills?q=" + q)
	if e != nil {
		return err("search failed")
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed")
}

	return ok("search completed")
}

func HandleInstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	_ = id
	resp, e := http.DefaultClient.Post("https://api.agentpowers.com/skills/install", "application/json", nil)
	if e != nil {
		return err("install failed")
}

	defer resp.Body.Close()
	return ok("installed")
}
