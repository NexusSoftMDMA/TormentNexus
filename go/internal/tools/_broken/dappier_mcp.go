package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func HandleDappierRealTimeSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "api_key")
	modelID, _ :=getString(args, "model_id")

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.dappier.com/search?query="+query+"&model_id="+modelID, nil)
	if e != nil {
		return err("error creating request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := http.DefaultClient
	resp, e := client.Do(req)
	if e != nil {
		return err("error making request: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err("server returned non-200 status: " + resp.Status)
}

	var response struct {
		Message string `json:"message"`
	}
e = json.NewDecoder(resp.Body).Decode(&response)
	if e != nil {
		return err("error decoding response: " + e.Error())
}

	return ok(response.Message)
}

var Manifest = struct {
	Filename   string `json:"filename"`
	ServerName string `json:"server_name"`
	Handlers   []struct {
		ToolName    string `json:"tool_name"`
		HandlerFunc string `json:"handler_func"`
		Description string `json:"description"`
	} `json:"handlers"`
}{
	Filename:   "dappier_mcp.go",
	ServerName: "dappier_mcp",
	Handlers: []struct {
		ToolName    string `json:"tool_name"`
		HandlerFunc string `json:"handler_func"`
		Description string `json:"description"`
	}{
		{
			ToolName:    "dappier_real_time_search",
			HandlerFunc: "HandleDappierRealTimeSearch",
			Description: "Retrieve real-time search data from Dappier by processing an AI model that supports two key capabilities: Real-Time Web Search and Stock Market Data.",
		},
	},
}