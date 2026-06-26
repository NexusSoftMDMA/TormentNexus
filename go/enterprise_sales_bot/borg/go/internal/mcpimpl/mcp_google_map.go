package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchPlaces_mcp_google_map(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "apiKey")
	if query == "" || apiKey == "" {
		return err("query and apiKey are required")
}

	u := fmt.Sprintf("https://maps.googleapis.com/maps/api/place/textsearch/json?query=%s&key=%s", url.QueryEscape(query), apiKey)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return ok(fmt.Sprintf("Places: %s", string(body)))
}

func HandleGetDirections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	origin, _ :=getString(args, "origin")
	dest, _ :=getString(args, "destination")
	apiKey, _ :=getString(args, "apiKey")
	if origin == "" || dest == "" || apiKey == "" {
		return err("origin, destination, and apiKey are required")
}

	u := fmt.Sprintf("https://maps.googleapis.com/maps/api/directions/json?origin=%s&destination=%s&key=%s", url.QueryEscape(origin), url.QueryEscape(dest), apiKey)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(fmt.Sprintf("Directions: %s", string(body)))
}