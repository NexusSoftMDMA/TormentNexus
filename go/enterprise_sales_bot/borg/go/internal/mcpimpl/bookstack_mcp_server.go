package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListShelves(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "api_url")
	tokenID, _ :=getString(args, "token_id")
	tokenSecret, _ :=getString(args, "token_secret")
	if baseURL == "" || tokenID == "" || tokenSecret == "" {
		return err("missing api_url, token_id, or token_secret")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/shelves", nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.SetBasicAuth(tokenID, tokenSecret)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json error: %v", e))
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleListBooks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "api_url")
	tokenID, _ :=getString(args, "token_id")
	tokenSecret, _ :=getString(args, "token_secret")
	if baseURL == "" || tokenID == "" || tokenSecret == "" {
		return err("missing api_url, token_id, or token_secret")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/books", nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.SetBasicAuth(tokenID, tokenSecret)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json error: %v", e))
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}