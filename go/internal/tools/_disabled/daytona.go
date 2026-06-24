package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleCreateWorkspace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "serverUrl")
	key, _ :=getString(args, "apiKey")
	body := map[string]string{"name": getString(args, "name")}
	b, e := json.Marshal(body)
	if e != nil {
		return err("marshal failed")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", url+"/workspaces", bytes.NewReader(b))
	if e != nil {
		return err("request failed")
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: "+e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err("bad status")
	}
	return ok("workspace created")
}

func HandleListWorkspaces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "serverUrl")
	key, _ :=getString(args, "apiKey")
	req, e := http.NewRequestWithContext(ctx, "GET", url+"/workspaces", nil)
	if e != nil {
		return err("request failed")
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: "+e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("bad status")
	}
	return ok("workspaces listed")
}