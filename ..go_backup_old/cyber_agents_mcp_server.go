package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSecurityScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		return err("target is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://scan.example.com/"+target, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("scan request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	return ok(fmt.Sprintf("scan completed: %v", result))
}

func HandleCheckConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configFile, found := args["config"].(string)
	if !found {
		return err("config must be a string")
}

	if configFile == "" {
		return err("config path is required")
}

	return ok(fmt.Sprintf("config %s is valid", configFile))
}