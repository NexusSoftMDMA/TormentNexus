package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateTestCase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseUrl, _ :=getString(args, "baseUrl")
	projectKey, _ :=getString(args, "projectKey")
	name, _ :=getString(args, "name")
	if baseUrl == "" || projectKey == "" || name == "" {
		return err("baseUrl, projectKey, and name are required")
}

	body := map[string]interface{}{"name": name, "objective": getString(args, "objective"), "status": getString(args, "status")}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request body")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, baseUrl+"/testcases", bytes.NewReader(jsonBody))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response body")
}

	var result map[string]interface{}
	if e = json.Unmarshal(respBody, &result); e != nil {
		return err("failed to parse response")
}

	key, found := result["key"].(string)
	if !found {
		return err("unexpected response format")
}

	return ok(fmt.Sprintf("Test case created successfully: %s", key))
}

func HandleGetTestCase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseUrl, _ :=getString(args, "baseUrl")
	projectKey, _ :=getString(args, "projectKey")
	testCaseKey, _ :=getString(args, "testCaseKey")
	if baseUrl == "" || projectKey == "" || testCaseKey == "" {
		return err("baseUrl, projectKey, and testCaseKey are required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, baseUrl+"/testcases/"+testCaseKey, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response body")
}

	var result map[string]interface{}
	if e = json.Unmarshal(respBody, &result); e != nil {
		return err("failed to parse response")
}

	return ok(fmt.Sprintf("Test case data: %s", string(respBody)))
}