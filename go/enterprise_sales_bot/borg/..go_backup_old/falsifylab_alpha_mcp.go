package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	service, _ :=getString(args, "service")
	if service == "" {
		service = "alpha"
	}
	url := fmt.Sprintf("https://api.falsifylab.com/%s/status", service)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	data, found := result["data"]
	if !found {
		data = result
	}
	return success(fmt.Sprintf("Status: %v", data))
}

func HandleGetAlphaData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		query = "latest"
	}
	url := fmt.Sprintf("https://api.falsifylab.com/alpha/data?q=%s", query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}