package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleCheckPolicy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	policyID, _ :=getString(args, "policy_id")
	baseURL, _ :=getString(args, "base_url")
	if policyID == "" {
		return err("policy_id is required")
}

	if baseURL == "" {
		baseURL = "https://api.mergeguide.com"
	}
	url := baseURL + "/policies/" + policyID
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok("Policy check result: " + string(body))
}