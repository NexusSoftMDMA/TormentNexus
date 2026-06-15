package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetMonitors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	key, _ :=getString(args, "api_key")
	if base == "" || key == "" {
		return err("Missing base_url or api_key")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/monitors", nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("X-API-Key", key)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("HTTP %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Read error: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("JSON error: " + e.Error())
}

	return ok(result)
}

func HandleGetGroups(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	key, _ :=getString(args, "api_key")
	if base == "" || key == "" {
		return err("Missing base_url or api_key")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/groups", nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("X-API-Key", key)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("HTTP %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Read error: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("JSON error: " + e.Error())
}

	return ok(result)
}