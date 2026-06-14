package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListContacts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	baseURL, _ :=getString(args, "base_url")
	url := fmt.Sprintf("%s/contacts?api_key=%s", baseURL, apiKey)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data interface{}
	e = json.Unmarshal(body, &data)
	if e != nil {
		return err("failed to parse response")
}

	b, _ := json.MarshalIndent(data, "", "  ")
	return success(string(b))
}

func HandleGetContact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	baseURL, _ :=getString(args, "base_url")
	id, _ :=getString(args, "id")
	url := fmt.Sprintf("%s/contacts/%s?api_key=%s", baseURL, id, apiKey)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data interface{}
	e = json.Unmarshal(body, &data)
	if e != nil {
		return err("failed to parse response")
}

	b, _ := json.MarshalIndent(data, "", "  ")
	return success(string(b))
}