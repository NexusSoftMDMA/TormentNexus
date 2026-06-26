package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cloudName, _ :=getString(args, "cloud_name")
	apiKey, _ :=getString(args, "api_key")
	apiSecret, _ :=getString(args, "api_secret")
	maxResults, _ :=getInt(args, "max_results")
	if maxResults == 0 {
		maxResults = 10
	}
	url := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/resources/image?max_results=%d", cloudName, maxResults)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
	}
	req.SetBasicAuth(apiKey, apiSecret)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
	}
	return ok(fmt.Sprintf("Resources: %v", result))
}

func HandleDeleteResource(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cloudName, _ :=getString(args, "cloud_name")
	apiKey, _ :=getString(args, "api_key")
	apiSecret, _ :=getString(args, "api_secret")
	publicID, _ :=getString(args, "public_id")
	url := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/resources/image/upload/%s", cloudName, publicID)
	req, e := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if e != nil {
		return err("failed to create request")
	}
	req.SetBasicAuth(apiKey, apiSecret)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
	}
	return success(fmt.Sprintf("Deleted: %v", result))
}