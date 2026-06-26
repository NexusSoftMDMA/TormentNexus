package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandlePatchSubscription(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	plan, _ :=getString(args, "plan")
	if token == "" || plan == "" {
		return err("token and plan are required")
}

	payload := map[string]string{
		"token": token,
		"plan":  plan,
		"scope": "claude-code:subscription",
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "PATCH", "https://api.anthropic.com/v1/oauth/subscription", strings.NewReader(string(body)))
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

	return ok(string(respBody))
}

func HandleGetSubscriptionStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.anthropic.com/v1/oauth/subscription?token="+token, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(respBody))
}