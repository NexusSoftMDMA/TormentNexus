package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func HandleCallService_mcp_server_home_assistant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	service, _ :=getString(args, "service")
	if domain == "" || service == "" {
		return err("domain and service are required")
}

	haURL := os.Getenv("HA_URL")
	haToken := os.Getenv("HA_TOKEN")
	if haURL == "" || haToken == "" {
		return err("HA_URL and HA_TOKEN environment variables required")
}

	entityID, _ :=getString(args, "entity_id")
	serviceDataRaw, _ :=getString(args, "service_data")
	var bodyMap map[string]interface{}
	if entityID != "" {
		bodyMap = map[string]interface{}{"entity_id": entityID}
	} else {
		bodyMap = map[string]interface{}{}
	}
	if serviceDataRaw != "" {
		var extra map[string]interface{}
		if e := json.Unmarshal([]byte(serviceDataRaw), &extra); e == nil {
			for k, v := range extra {
				bodyMap[k] = v
			}
		}
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	url := fmt.Sprintf("%s/api/services/%s/%s", strings.TrimRight(haURL, "/"), domain, service)
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyBytes)))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+haToken)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBytes, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Home Assistant error %d: %s", resp.StatusCode, string(respBytes)))
}

	return ok(fmt.Sprintf("Service %s.%s called successfully", domain, service))
}