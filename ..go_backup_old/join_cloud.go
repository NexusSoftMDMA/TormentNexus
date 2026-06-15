package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleJoinCloudCreate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	body := map[string]string{"name": name}
	data, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.join.cloud/v1/deployments", bytes.NewReader(data))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err("API returned status " + resp.Status)
}

	return success("Deployment created: " + name)
}