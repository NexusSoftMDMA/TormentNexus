//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetFlag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectKey, _ :=getString(args, "project_key")
	flagKey, _ :=getString(args, "flag_key")
	if projectKey == "" || flagKey == "" {
		return err("project_key and flag_key are required")
}

	apiKey := os.Getenv("LAUNCHDARKLY_API_KEY")
	if apiKey == "" {
		return err("LAUNCHDARKLY_API_KEY not set")
}

	url := fmt.Sprintf("https://app.launchdarkly.com/api/v2/flags/%s/%s", projectKey, flagKey)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", api


-reasoner (deepseek)*
}