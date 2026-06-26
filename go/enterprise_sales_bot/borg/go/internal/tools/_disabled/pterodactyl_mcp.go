package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	baseURL, _ :=getString(args, "base_url")
	if apiKey == "" || baseURL == "" {
		return err("missing api_key or base_url")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/application/servers", nil)
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json error: " + e.Error())
}

	// Now we can use 'found' for type assertion if needed, but we just return result
	// For brevity, just return the raw data as a string
	data, _ := json.Marshal(result) // ignore error for simplicity
	return ok(string(data))
}

func HandleGetServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	baseURL, _ :=getString(args, "base_url")
	serverID, _ :=getString(args, "server_id")
	if apiKey == "" || baseURL == "" || serverID == "" {
		return err("missing api_key, base_url, or server_id")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/application/servers/"+serverID, nil)
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json error: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}