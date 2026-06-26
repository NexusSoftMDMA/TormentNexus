package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// HandleListEntities lists New Relic entities.
func HandleListEntities(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.newrelic.com/v2/entities.json", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Api-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
	}
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response: " + e.Error())
	}
	entities, found := result["entities"].([]interface{})
	if !found {
		return err("no entities in response")
	}
	return ok(fmt.Sprintf("Found %d entities", len(entities)))
}

// HandleGetEntity gets a specific entity by ID.
func HandleGetEntity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	entityID, _ :=getString(args, "entityId")
	if apiKey == "" || entityID == "" {
		return err("apiKey and entityId are required")
	}
	u, e := url.Parse("https://api.newrelic.com/v2/entities/" + entityID + ".json")
	if e != nil {
		return err("invalid URL")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Api-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
	}
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response: " + e.Error())
	}
	return ok(fmt.Sprintf("Entity: %+v", result))
}