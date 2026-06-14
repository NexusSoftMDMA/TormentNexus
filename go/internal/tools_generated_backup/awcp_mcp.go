package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleDelegate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")
	amount, _ :=getString(args, "amount")
	body := map[string]string{"from": from, "to": to, "amount": amount}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal body: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode >= 400 {
		return err("server error: " + string(respBody))
}

	return ok("delegation successful: " + string(respBody))
}

func HandleRevokeDelegation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode >= 400 {
		return err("server error: " + string(respBody))
}

	return ok("delegation revoked: " + string(respBody))
}