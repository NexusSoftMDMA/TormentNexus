package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListProperties(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	location, _ :=getString(args, "location")
	if location == "" {
		return err("location is required")
}

	url := fmt.Sprintf("https://api.localsbnb.com/properties?location=%s", location)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch properties: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result []map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d properties", len(result)))
}

func HandleGetProperty(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := fmt.Sprintf("https://api.localsbnb.com/properties/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch property: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Property: %v", result))
}