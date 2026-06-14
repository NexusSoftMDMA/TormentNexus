package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetIntegration(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "integration_id")
	if id == "" {
		return err("integration_id is required")
}

	url := fmt.Sprintf("https://api.snaprender.com/integrations/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch integration: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Integration: %v", result))
}

func HandleListIntegrations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.snaprender.com/integrations"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list integrations: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result []interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Integrations: %v", result))
}