package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
)

func HandleAskExpert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	expertise, _ :=getString(args, "expertise")
	payload, _ := json.Marshal(map[string]string{"query": query, "expertise": expertise})
	apiURL := os.Getenv("AI_API_URL")
	if apiURL == "" {
		apiURL = "https://api.example.com/ask"
	}
	req, e := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("AI_API_KEY"))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	answer, found := result["answer"].(string)
	if !found {
		return err("no answer in response")
}

	return success(answer)
}

func HandleHandoff(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	contextData, _ :=getString(args, "context")
	if contextData == "" {
		return err("context is required")
}

	handoffID := "handoff-" + contextData[:min(len(contextData), 8)]
	return ok("Handoff created with ID: " + handoffID)
}