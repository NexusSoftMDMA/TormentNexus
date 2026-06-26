package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleMineru(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("request creation failed")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("non-200 status")
}

	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed")
}

	return success("miner data retrieved")
}

func HandleMineruStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "status_url")
	if url == "" {
		return err("missing status_url")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("status request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("non-200 status")
}

	var status map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&status); e != nil {
		return err("decode failed")
}

	found, _ := status["online"].(bool)
	if !found {
		return ok("status unknown")
}

	if found {
		return success("miner online")
}

	return ok("miner offline")
}// touch 1781132135
