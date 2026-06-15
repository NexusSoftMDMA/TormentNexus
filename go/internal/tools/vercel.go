//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// callVercelAPI makes requests to the Vercel REST API.
func callVercelAPI(ctx context.Context, method string, urlPath string, queryParams map[string]string, body interface{}) ([]byte, error) {
	vercelToken := os.Getenv("VERCEL_TOKEN")
	if vercelToken == "" {
		return nil, fmt.Errorf("VERCEL_TOKEN environment variable is not set")
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	baseUrl := os.Getenv("VERCEL_API_URL")
	if baseUrl == "" {
		baseUrl = "https://api.vercel.com"
	}
	reqUrl := baseUrl + urlPath

	req, err := http.NewRequestWithContext(ctx, method, reqUrl, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+vercelToken)
	req.Header.Set("Content-Type", "application/json")

	if len(queryParams) > 0 {
		q := req.URL.Query()
		for k, v := range queryParams {
			if v != "" {
				q.Add(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("vercel API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// HandleVercelListProjects lists all projects in Vercel.
func HandleVercelListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	queryParams := make(map[string]string)
	if limit := getInt(args, "limit"); limit > 0 {
		queryParams["limit"] = strconv.Itoa(limit)
	}
	if since := getInt(args, "since"); since > 0 {
		queryParams["since"] = strconv.Itoa(since)
	}
	if until := getInt(args, "until"); until > 0 {
		queryParams["until"] = strconv.Itoa(until)
	}

	respData, err := callVercelAPI(ctx, "GET", "/v9/projects", queryParams, nil)
	if err != nil {
		return errResponseVercel(err)
	}

	return ok(string(respData))
}

// HandleVercelGetProject gets details for a specific project.
func HandleVercelGetProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectId, _ := getString(args, "projectId")
	if projectId == "" {
		return err("projectId parameter is required")
	}

	urlPath := fmt.Sprintf("/v9/projects/%s", projectId)
	respData, err := callVercelAPI(ctx, "GET", urlPath, nil, nil)
	if err != nil {
		return errResponseVercel(err)
	}

	return ok(string(respData))
}

// HandleVercelListDeployments lists Vercel deployments.
func HandleVercelListDeployments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	queryParams := make(map[string]string)
	if projectId, _ := getString(args, "projectId"); projectId != "" {
		queryParams["projectId"] = projectId
	}
	if limit := getInt(args, "limit"); limit > 0 {
		queryParams["limit"] = strconv.Itoa(limit)
	}
	if since := getInt(args, "since"); since > 0 {
		queryParams["since"] = strconv.Itoa(since)
	}
	if until := getInt(args, "until"); until > 0 {
		queryParams["until"] = strconv.Itoa(until)
	}
	if state, _ := getString(args, "state"); state != "" {
		queryParams["state"] = state
	}

	respData, err := callVercelAPI(ctx, "GET", "/v6/deployments", queryParams, nil)
	if err != nil {
		return errResponseVercel(err)
	}

	return ok(string(respData))
}

// HandleVercelGetDeployment gets a specific deployment details.
func HandleVercelGetDeployment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deploymentId, _ := getString(args, "deploymentId")
	if deploymentId == "" {
		return err("deploymentId parameter is required")
	}

	urlPath := fmt.Sprintf("/v13/deployments/%s", deploymentId)
	respData, err := callVercelAPI(ctx, "GET", urlPath, nil, nil)
	if err != nil {
		return errResponseVercel(err)
	}

	return ok(string(respData))
}

// HandleVercelCancelDeployment cancels a running deployment.
func HandleVercelCancelDeployment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deploymentId, _ := getString(args, "deploymentId")
	if deploymentId == "" {
		return err("deploymentId parameter is required")
	}

	urlPath := fmt.Sprintf("/v12/deployments/%s/cancel", deploymentId)
	respData, err := callVercelAPI(ctx, "POST", urlPath, nil, nil)
	if err != nil {
		return errResponseVercel(err)
	}

	return ok(string(respData))
}

// HandleVercelListEnvVars lists environment variables for a project.
func HandleVercelListEnvVars(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectId, _ := getString(args, "projectId")
	if projectId == "" {
		return err("projectId parameter is required")
	}

	urlPath := fmt.Sprintf("/v9/projects/%s/env", projectId)
	respData, err := callVercelAPI(ctx, "GET", urlPath, nil, nil)
	if err != nil {
		return errResponseVercel(err)
	}

	return ok(string(respData))
}

// HandleVercelCreateEnvVar creates an environment variable.
func HandleVercelCreateEnvVar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectId, _ := getString(args, "projectId")
	if projectId == "" {
		return err("projectId parameter is required")
	}

	key, _ := getString(args, "key")
	if key == "" {
		return err("key parameter is required")
	}

	value, _ := getString(args, "value")
	if value == "" {
		return err("value parameter is required")
	}

	target := []string{"development", "preview", "production"}
	if targetVal, exists := args["target"]; exists {
		if rawArray, ok := targetVal.([]interface{}); ok {
			var parsedTarget []string
			for _, item := range rawArray {
				if s, okS := item.(string); okS {
					parsedTarget = append(parsedTarget, s)
				}
			}
			if len(parsedTarget) > 0 {
				target = parsedTarget
			}
		}
	}

	body := map[string]interface{}{
		"key":    key,
		"value":  value,
		"type":   "plain",
		"target": target,
	}

	urlPath := fmt.Sprintf("/v10/projects/%s/env", projectId)
	respData, err := callVercelAPI(ctx, "POST", urlPath, nil, body)
	if err != nil {
		return errResponseVercel(err)
	}

	return ok(string(respData))
}

// HandleVercelDeleteEnvVar deletes an environment variable.
func HandleVercelDeleteEnvVar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectId, _ := getString(args, "projectId")
	if projectId == "" {
		return err("projectId parameter is required")
	}

	envVarId, _ := getString(args, "envVarId")
	if envVarId == "" {
		return err("envVarId parameter is required")
	}

	urlPath := fmt.Sprintf("/v9/projects/%s/env/%s", projectId, envVarId)
	respData, err := callVercelAPI(ctx, "DELETE", urlPath, nil, nil)
	if err != nil {
		return errResponseVercel(err)
	}

	return ok(string(respData))
}

func errResponseVercel(err error) (ToolResponse, error) {
	return ToolResponse{
		Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Vercel API call failed: %v", err)}},
		IsError: true,
	}, nil
}
