package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// getBaseURL returns the Heurist API base URL from environment or default.
func getBaseURL() string {
	base := os.Getenv("HEURIST_API_BASE_URL")
	if base == "" {
		base = "http://localhost:8000"
	}
	return strings.TrimRight(base, "/")
}

// callAPI performs an HTTP request to the Heurist API.
func callAPI(method, path, body string) ([]byte, error) {
	client := http.DefaultClient
	url := getBaseURL() + "/" + strings.TrimLeft(path, "/")

	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)

	req, reqErr := http.NewRequest(method, url, reqBody)
	if reqErr != nil {
		return nil, fmt.Errorf("failed to create request: %w", reqErr)
}

	req.Header.Set("Accept", "application/json")
	if body != "" {
		req.Header.Set("Content-Type", "application/json")

	resp, apiErr := client.Do(req)
	if apiErr != nil {
		return nil, fmt.Errorf("API request failed: %w", apiErr)
}

	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", readErr)
}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
}

	return respBody, nil
}

}
}

// HandleListAgents returns a list of available agents.
func HandleListAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	body, apiErr := callAPI("GET", "agents", "")
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(string(body))
}

// HandleGetAgent returns details for a specific agent.
func HandleGetAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agentID, _ :=getString(args, "agent_id")
	if agentID == "" {
		return err("agent_id is required")
}

	body, apiErr := callAPI("GET", "agents/"+agentID, "")
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(string(body))
}

// HandleRunAgent executes an agent with the given input.
func HandleRunAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agentID, _ :=getString(args, "agent_id")
	if agentID == "" {
		return err("agent_id is required")
}

	input, _ :=getString(args, "input")
	if input == "" {
		return err("input is required")
}

	payload := fmt.Sprintf(`{"input":%s}`, mustMarshal(input))
	body, apiErr := callAPI("POST", "agents/"+agentID+"/run", payload)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(string(body))
}

// mustMarshal serializes v to JSON string, panics on failure.
func mustMarshal(v interface{}) string {
	b, e := json.Marshal(v)
	if e != nil {
		// This should never happen with simple types
		panic("json marshal failed: " + e.Error())

	return string(b)
}
}