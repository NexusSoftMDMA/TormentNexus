package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetAccounts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "accessToken")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://gateway.saxobank.com/sim/openapi/port/v1/accounts/me", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	return ok(string(body))
}

func HandleGetAccountSummary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "accessToken")
	accountKey, _ :=getString(args, "accountKey")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://gateway.saxobank.com/sim/openapi/port/v1/accounts/"+accountKey+"/summary", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	return ok(string(body))
}