package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchPhotos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	u, _ := url.Parse("https://api.unsplash.com/search/photos")
	u.RawQuery = url.Values{"query": {query}, "client_id": {apiKey}}.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json error: " + e.Error())
}

	if resp.StatusCode != 200 {
		msg, _ := result["errors"].([]interface{})
		return err(fmt.Sprintf("API error: %v", msg))
}

	return ok(fmt.Sprintf("Found %d photos", int(result["total"].(float64))))
}

func HandleGetPhoto(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id parameter is required")
}

	u := fmt.Sprintf("https://api.unsplash.com/photos/%s?client_id=%s", id, apiKey)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json error: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %v", result["errors"]))
}

	return ok(fmt.Sprintf("Photo: %s by %s", result["id"], result["user"].(map[string]interface{})["name"]))
}