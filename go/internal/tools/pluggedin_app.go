//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiURL, _ :=getString(args, "api_url")
	if apiURL == "" {
		return err("api_url is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/servers", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to execute request: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e


-reasoner (deepseek)*
}