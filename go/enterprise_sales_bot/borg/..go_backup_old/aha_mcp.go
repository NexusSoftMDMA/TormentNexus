package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type graphqlRequest struct {
	Query string `json:"query"`
}

type graphqlResponse struct {
	Data json.RawMessage `json:"data"`
	Errors []struct{Message string} `json:"errors"`
}

func HandleListFeatures(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("AHA_API_KEY")
	if apiKey == "" {
		return err("AHA_API_KEY not set")
}

	query := `{ features { nodes { id name } } }`
	body, e := json.Marshal(graphqlRequest{Query: query})
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://yourdomain.aha.io/api/v1/graphql", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var gqlResp graphqlResponse
	if e = json.Unmarshal(respBody, &gqlResp); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(gqlResp.Errors) > 0 {
		return err(gqlResp.Errors[0].Message)
}

	return ok(string(gqlResp.Data))
}

func HandleListIdeas(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("AHA_API_KEY")
	if apiKey == "" {
		return err("AHA_API_KEY not set")
}

	query := `{ ideas { nodes { id name } } }`
	body, e := json.Marshal(graphqlRequest{Query: query})
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://yourdomain.aha.io/api/v1/graphql", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var gqlResp graphqlResponse
	if e = json.Unmarshal(respBody, &gqlResp); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(gqlResp.Errors) > 0 {
		return err(gqlResp.Errors[0].Message)
}

	return ok(string(gqlResp.Data))
}