package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleListServices_gogcli_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	region, _ :=getString(args, "region")
	url := "https://www.googleapis.com/discovery/v1/apis"
	if region != "" {
		url += "?region=" + region
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch services: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	items, found := result["items"].([]interface{})
	if !found {
		return err("missing items in response")
}

	names := make([]string, 0, len(items))
	for _, item := range items {
		m, found := item.(map[string]interface{})
		if found {
			if name, found := m["name"].(string); found {
				names = append(names, name)

		}
	}
	return ok(strings.Join(names, ", "))
}

}

func HandleDescribeService(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	service, _ :=getString(args, "service")
	if service == "" {
		return err("service argument is required")
}

	url := "https://www.googleapis.com/discovery/v1/apis/" + service + "/rest"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch service: " + e.Error())
}

	defer resp.Body.Close()
	var doc map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&doc); e != nil {
		return err("failed to decode response: " + e.Error())
}

	title, _ := doc["title"].(string)
	return success(title)
}