package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListApis(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.myapi.com/apis", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var apis []string
	if e := json.Unmarshal(body, &apis); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Available APIs: %v", apis))
}

func HandleCallApi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiName, _ :=getString(args, "apiName")
	endpoint, _ :=getString(args, "endpoint")
	if apiName == "" || endpoint == "" {
		return err("apiName and endpoint are required")
}

	url := fmt.Sprintf("https://api.myapi.com/apis/%s/%s", apiName, endpoint)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}