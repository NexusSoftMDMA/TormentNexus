package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetVerdict(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
	}
	url := fmt.Sprintf("https://api.verdictswarm.dev/verdict?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
	}
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
	}
	return success(fmt.Sprintf("Verdict: %v", result["verdict"]))
}