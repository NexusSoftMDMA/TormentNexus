package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListIntegrations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.example.com/integrations")
	if e != nil {
		return err("failed to fetch integrations: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(fmt.Sprintf("Integrations: %v", result))
}

func HandleGetIntegration(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := "https://api.example.com/integrations/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch integration: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(fmt.Sprintf("Integration %s: %v", id, result))
}