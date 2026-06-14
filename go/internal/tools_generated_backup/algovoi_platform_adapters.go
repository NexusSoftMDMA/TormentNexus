package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListAdapters(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.algovoi.com/v1/adapters", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	var result []map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d adapters", len(result)))
}

func HandleGetAdapter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	adapterID, _ :=getString(args, "adapter_id")
	if adapterID == "" {
		return err("adapter_id is required")
}

	url := fmt.Sprintf("https://api.algovoi.com/v1/adapters/%s", adapterID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Adapter: %v", result["name"]))
}