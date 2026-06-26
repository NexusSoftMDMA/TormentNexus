package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleDeployApp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "api_token")
	appID, _ :=getString(args, "app_id")
	if token == "" || appID == "" {
		return err("api_token and app_id are required")
}

	url := "https://api.hostman.com/v1/apps/" + appID + "/deploy"
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		var msg map[string]interface{}
		json.Unmarshal(body, &msg)
		return err(fmt.Sprintf("deploy failed (%d): %v", resp.StatusCode, msg))
}

	return success("deployment triggered successfully")
}