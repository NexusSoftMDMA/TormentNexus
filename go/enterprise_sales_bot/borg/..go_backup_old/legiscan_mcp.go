package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleLegiscanSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	state, _ :=getString(args, "state")
	if query == "" {
		return err("query parameter is required")
}

	apiKey := os.Getenv("LEGISCAN_API_KEY")
	if apiKey == "" {
		return err("LEGISCAN_API_KEY not set")
}

	url := fmt.Sprintf("https://api.legiscan.com/?key=%s&op=getSearch&state=%s&query=%s", apiKey, state, query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result struct {
		Status string `json:"status"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse failed: %v", e))
}

	return ok(fmt.Sprintf("Search results: %s", string(body)))
}