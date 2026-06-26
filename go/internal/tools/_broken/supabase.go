package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const supabaseAPIURL = "https://api.supabase.com/v1"

func HandleSupabaseQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectRef, _ :=getString(args, "project_ref")
	table, _ :=getString(args, "table")
	apiKey, _ :=getString(args, "api_key")

	if projectRef == "" || table == "" || apiKey == "" {
		return err("project_ref, table, and api_key are required")
}

	query, _ :=getString(args, "query")
	filters, _ :=getString(args, "filters")

	urlStr := fmt.Sprintf("%s/%s/rest/v1/%s", supabaseAPIURL, projectRef, table)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	if query != "" {
		req.URL.RawQuery = query
	}

	if filters != "" {
		if req.URL.RawQuery != "" {
			req.URL.RawQuery += "&" + filters
		} else {
			req.URL.RawQuery = filters
		}
	}

	client := http.DefaultClient
	resp, apiErr := client.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&result)
	if decodeErr != nil {
		return err(fmt.Sprintf("failed to decode response: %v", decodeErr))
}

	jsonData, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal response: %v", jsonErr))
}

	return ok(string(jsonData))
}

func HandleSupabaseInsert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectRef, _ :=getString(args, "project_ref")
	table, _ :=getString(args, "table")
	apiKey, _ :=getString(args, "api_key")
	data, _ :=getString(args, "data")

	if projectRef == "" || table == "" || apiKey == "" || data == "" {
		return err("project_ref, table, api_key, and data are required")
}

	urlStr := fmt.Sprintf("%s/%s/rest/v1/%s", supabaseAPIURL, projectRef, table)

	req, reqErr := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(data))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	client := http.DefaultClient
	resp, apiErr := client.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&result)
	if decodeErr != nil {
		return err(fmt.Sprintf("failed to decode response: %v", decodeErr))
}

	jsonData, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal response: %v", jsonErr))
}

	return ok(string(jsonData))
}

func HandleSupabaseUpdate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectRef, _ :=getString(args, "project_ref")
	table, _ :=getString(args, "table")
	apiKey, _ :=getString(args, "api_key")
	data, _ :=getString(args, "data")
	match, _ :=getString(args, "match")

	if projectRef == "" || table == "" || apiKey == "" || data == "" || match == "" {
		return err("project_ref, table, api_key, data, and match are required")
}

	urlStr := fmt.Sprintf("%s/%s/rest/v1/%s?%s", supabaseAPIURL, projectRef, table, match)

	req, reqErr := http.NewRequestWithContext(ctx, "PATCH", urlStr, strings.NewReader(data))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	client := http.DefaultClient
	resp, apiErr := client.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&result)
	if decodeErr != nil {
		return err(fmt.Sprintf("failed to decode response: %v", decodeErr))
}

	jsonData, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal response: %v", jsonErr))
}

	return ok(string(jsonData))
}

func HandleSupabaseDelete(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectRef, _ :=getString(args, "project_ref")
	table, _ :=getString(args, "table")
	apiKey, _ :=getString(args, "api_key")
	match, _ :=getString(args, "match")

	if projectRef == "" || table == "" || apiKey == "" || match == "" {
		return err("project_ref, table, api_key, and match are required")
}

	urlStr := fmt.Sprintf("%s/%s/rest/v1/%s?%s", supabaseAPIURL, projectRef, table, match)

	req, reqErr := http.NewRequestWithContext(ctx, "DELETE", urlStr, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, apiErr := client.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	return ok("Successfully deleted records")
}

func HandleSupabaseAuth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectRef, _ :=getString(args, "project_ref")
	apiKey, _ :=getString(args, "api_key")
	email, _ :=getString(args, "email")
	password, _ :=getString(args, "password")

	if projectRef == "" || apiKey == "" || email == "" || password == "" {
		return err("project_ref, api_key, email, and password are required")
}

	urlStr := fmt.Sprintf("%s/%s/auth/v1/token?grant_type=password", supabaseAPIURL, projectRef)

	data := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	req, reqErr := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(data))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, apiErr := client.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("Auth error: %s - %s", resp.Status, string(body)))
}

	var result interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&result)
	if decodeErr != nil {
		return err(fmt.Sprintf("failed to decode response: %v", decodeErr))
}

	jsonData, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal response: %v", jsonErr))
}

	return ok(string(jsonData))
}