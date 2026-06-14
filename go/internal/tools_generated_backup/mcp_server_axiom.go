package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleAxiomQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dataset, _ :=getString(args, "dataset")
	query, _ :=getString(args, "query")
	if dataset == "" || query == "" {
		return err("dataset and query are required")
}

	apiKey := os.Getenv("AXIOM_TOKEN")
	if apiKey == "" {
		return err("AXIOM_TOKEN not set")
}

	baseURL := os.Getenv("AXIOM_URL")
	if baseURL == "" {
		baseURL = "https://api.axiom.co"
	}
	url := fmt.Sprintf("%s/v1/datasets/%s/query", baseURL, dataset)
	body := map[string]string{"query": query}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(strings.NewReader(string(payload))))
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("decode error: " + e.Error())
}

	jsonBytes, _ := json.Marshal(result)
	return success(string(jsonBytes))
}

func HandleAxiomListDatasets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("AXIOM_TOKEN")
	if apiKey == "" {
		return err("AXIOM_TOKEN not set")
}

	baseURL := os.Getenv("AXIOM_URL")
	if baseURL == "" {
		baseURL = "https://api.axiom.co"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/v1/datasets", nil)
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("decode error: " + e.Error())
}

	jsonBytes, _ := json.Marshal(result)
	return success(string(jsonBytes))
}