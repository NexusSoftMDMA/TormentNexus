package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetReputation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("domain is required")
}

	url := fmt.Sprintf("https://api.example.com/reputation?domain=%s", domain)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch reputation: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	score, found := result["score"].(float64)
	if !found {
		return err("score not found in response")
}

	return success(fmt.Sprintf("Reputation score for %s: %.2f", domain, score))
}