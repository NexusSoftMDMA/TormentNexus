package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListVolumes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetIP, _ :=getString(args, "target_ip")
	apiToken, _ :=getString(args, "api_token")
	if targetIP == "" || apiToken == "" {
		return err("missing required parameters: target_ip and api_token")
}

	url := fmt.Sprintf("https://%s/api/1.16/volume", targetIP)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result []map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d volumes", len(result)))
}