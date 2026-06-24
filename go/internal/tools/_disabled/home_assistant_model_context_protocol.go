package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleListStates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := os.Getenv("HOME_ASSISTANT_URL")
	token := os.Getenv("HOME_ASSISTANT_TOKEN")
	if url == "" || token == "" {
		return err("HOME_ASSISTANT_URL and HOME_ASSISTANT_TOKEN must be set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url+"/api/states", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var states []map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&states)
	if e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("Found %d states", len(states)))
}

func HandleCallService(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := os.Getenv("HOME_ASSISTANT_URL")
	token := os.Getenv("HOME_ASSISTANT_TOKEN")
	if url == "" || token == "" {
		return err("HOME_ASSISTANT_URL and HOME_ASSISTANT_TOKEN must be set")
}

	domain, _ :=getString(args, "domain")
	service, _ :=getString(args, "service")
	entityID, _ :=getString(args, "entity_id")
	if domain == "" || service == "" {
		return err("domain and service are required")
}

	body := map[string]interface{}{}
	if entityID != "" {
		body["entity_id"] = entityID
	}
	b, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("marshal failed: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", url+"/api/services/"+domain+"/"+service, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success("Service call succeeded")
}