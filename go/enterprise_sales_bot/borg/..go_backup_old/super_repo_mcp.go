package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateFeature(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiURL, _ :=getString(args, "api_url")
	repo, _ :=getString(args, "repo")
	featureName, _ :=getString(args, "feature_name")
	if apiURL == "" || repo == "" || featureName == "" {
		return err("api_url, repo, and feature_name are required")
}

	body, _ := json.Marshal(map[string]string{"repo": repo, "feature_name": featureName})
	req, e := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(respBody)))
}

	return ok(fmt.Sprintf("Feature '%s' created in repo '%s'", featureName, repo))
}