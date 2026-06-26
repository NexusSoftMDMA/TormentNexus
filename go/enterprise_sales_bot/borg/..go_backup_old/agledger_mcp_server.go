package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleAccounts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.agledger.com/accounts", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(string(body))
}

}

func HandleTransaction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	payload := map[string]interface{}{
		"from":     getString(args, "from"),
		"to":       getString(args, "to"),
		"amount":   getString(args, "amount"),
		"currency": getString(args, "currency"),
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.agledger.com/transactions", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	apiKey, _ :=getString(args, "api_key")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(body))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(respBody))
}
}