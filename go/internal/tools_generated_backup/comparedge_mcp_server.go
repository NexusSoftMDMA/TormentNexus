package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleComparePricing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	toolName, _ :=getString(args, "tool")
	if toolName == "" {
		return err("tool parameter is required")
}

	url := "https://api.comparedge.com/pricing?tool=" + toolName
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch pricing: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}