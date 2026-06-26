package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListCollections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://api.binalyze.com"
	}
	url := fmt.Sprintf("%s/v1/collections", baseURL)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	data, found := result["data"]
	if !found {
		return err("data field missing in response")
}

	return success(fmt.Sprintf("Collections: %v", data))
}

func HandleGetCollection(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://api.binalyze.com"
	}
	collectionID, _ :=getString(args, "collection_id")
	if collectionID == "" {
		return err("collection_id is required")
}

	url := fmt.Sprintf("%s/v1/collections/%s", baseURL, collectionID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	data, found := result["data"]
	if !found {
		return err("data field missing in response")
}

	return success(fmt.Sprintf("Collection: %v", data))
}