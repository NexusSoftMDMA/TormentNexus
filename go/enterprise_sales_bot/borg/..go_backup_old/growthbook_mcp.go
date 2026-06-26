package tools

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
)

func HandleGetFeatureFlag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("GROWTHBOOK_API_KEY")
	key, _ :=getString(args, "key")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.growthbook.io/api/v1/features/"+key, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return ok(string(body))
}

func HandleListFeatures(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("GROWTHBOOK_API_KEY")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.growthbook.io/api/v1/features", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return ok(string(body))
}