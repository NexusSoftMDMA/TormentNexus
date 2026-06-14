package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleExploreMarketplace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://mcp-marketplace.example.com/api/servers"
	if v, found := args["url"].(string); found {
		url = v
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch marketplace")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response")
}

	return ok("Marketplace data: " + fmt.Sprintf("%v", result))
}