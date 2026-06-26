package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
)

func HandleMifosxListClients(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	url := fmt.Sprintf("%s/fineract-provider/api/v1/clients", baseURL)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("JSON parse error: %v", e))
}

	return ok(fmt.Sprintf("Clients: %s", string(body)))
}

func HandleMifosxGetClient(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	clientID, _ :=getString(args, "client_id")
	if baseURL == "" || clientID == "" {
		return err("base_url and client_id are required")
}

	url := fmt.Sprintf("%s/fineract-provider/api/v1/clients/%s", baseURL, clientID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d %s", resp.StatusCode, string(body)))
}

	return ok(fmt.Sprintf("Client details: %s", string(body)))
}