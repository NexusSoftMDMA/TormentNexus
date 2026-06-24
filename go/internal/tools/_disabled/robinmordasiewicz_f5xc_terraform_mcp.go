package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
)

func HandleListAPISpecs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.example.com/f5xc/specs")
	if e != nil {
		return err(fmt.Sprintf("failed to fetch specs: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return success(fmt.Sprintf("Found %d API specs", len(result)))
}

func HandleGetSubscriptionInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.example.com/f5xc/subscription")
	if e != nil {
		return err(fmt.Sprintf("failed to fetch subscription: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Subscription: %v", result["status"]))
}