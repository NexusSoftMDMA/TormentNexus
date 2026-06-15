package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetComponentType(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	url := fmt.Sprintf("https://api.donotdev.com/component-types/%s", name)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(result)
}

func HandleListComponentTypes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	url := "https://api.donotdev.com/component-types"
	if filter != "" {
		url += "?filter=" + filter
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	var result []interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(result)
}