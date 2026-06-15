package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetCandidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	candidateId, _ :=getString(args, "candidateId")
	if candidateId == "" {
		return err("candidateId is required")
}

	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
}

	url := fmt.Sprintf("https://api.kula.ai/v1/candidates/%s", candidateId)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("API status: " + resp.Status)
}

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	jsonBytes, e := json.Marshal(data)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	return success(string(jsonBytes))
}

func HandleListCandidates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
}

	url := "https://api.kula.ai/v1/candidates"
	if query != "" {
		url += "?q=" + query
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("API status: " + resp.Status)
}

	var data []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	jsonBytes, e := json.Marshal(data)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	return success(string(jsonBytes))
}