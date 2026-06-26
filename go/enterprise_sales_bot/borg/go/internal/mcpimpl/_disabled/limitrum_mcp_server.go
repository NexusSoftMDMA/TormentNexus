package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleVerifyPolicy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	policy, _ :=getString(args, "policy")
	data, _ :=getString(args, "data")
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	body, e := json.Marshal(map[string]string{"policy": policy, "data": data})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response: " + e.Error())
}

	msg, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result: " + e.Error())
}

	return ok("Policy verification result: " + string(msg))
}