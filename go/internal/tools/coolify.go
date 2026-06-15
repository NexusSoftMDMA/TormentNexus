//go:build ignore
// +build ignore

package tools

/**
 * @file coolify.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Coolify MCP — deployment and infrastructure management.
 * Replaces: github.com/coollabsio/coolify-mcp
 *
 * Provides deployment management, resource orchestration, and infrastructure
 * automation via the Coolify API (open-source PaaS alternative to Heroku/Netlify).
 * Configurable via COOLIFY_API_URL and COOLIFY_API_TOKEN env vars.
 *
 * Tools:
 *  - coolify_list_projects — list all projects
 *  - coolify_create_project — create a new project
 *  - coolify_list_services — list services in a project
 *  - coolify_deploy_service — deploy a service
 *  - coolify_get_logs — get service logs
 *  - coolify_list_databases — list database resources
 */

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func coolifyBaseURL() string {
	if u := os.Getenv("COOLIFY_API_URL"); u != "" {
		return u
	}
	return "https://app.coolify.io/api/v1"
}

func coolifyToken() string {
	return os.Getenv("COOLIFY_API_TOKEN")
}

func coolifyRequest(ctx context.Context, method, path string, payload map[string]interface{}) (string, error) {
	var bodyReader io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		bodyReader = bytes.NewReader(b)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, e := http.NewRequestWithContext(ctx, method, coolifyBaseURL()+path, bodyReader)
	if e != nil {
		return "", fmt.Errorf("request error: %v", e)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if token := coolifyToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, e := client.Do(req)
	if e != nil {
		return "", fmt.Errorf("Coolify API error: %v", e)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("Coolify error (%d): %s", resp.StatusCode, string(data))
	}
	return string(data), nil
}

// HandleCoolifyListProjects lists all projects.
func HandleCoolifyListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	result, e := coolifyRequest(ctx, "GET", "/projects", nil)
	if e != nil {
		return err(fmt.Sprintf("list projects failed: %v", e))
	}
	return ok(result)
}

// HandleCoolifyCreateProject creates a new project.
func HandleCoolifyCreateProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ := getString(args, "name")
	if name == "" {
		return err("name is required")
	}
	
	payload := map[string]interface{}{"name": name}
	if desc, _ := getString(args, "description"); desc != "" {
		payload["description"] = desc
	}

	result, e := coolifyRequest(ctx, "POST", "/projects", payload)
	if e != nil {
		return err(fmt.Sprintf("create project failed: %v", e))
	}
	return ok(result)
}

// HandleCoolifyListServices lists services in a project.
func HandleCoolifyListServices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectUUID, _ := getString(args, "project_uuid", "project_id")
	if projectUUID == "" {
		return err("project_uuid is required")
	}

	result, e := coolifyRequest(ctx, "GET", fmt.Sprintf("/projects/%s/services", projectUUID), nil)
	if e != nil {
		return err(fmt.Sprintf("list services failed: %v", e))
	}
	return ok(result)
}

// HandleCoolifyDeployService deploys a service.
func HandleCoolifyDeployService(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serviceUUID, _ := getString(args, "service_uuid", "service_id")
	if serviceUUID == "" {
		return err("service_uuid is required")
	}

	result, e := coolifyRequest(ctx, "POST", fmt.Sprintf("/services/%s/deploy", serviceUUID), nil)
	if e != nil {
		return err(fmt.Sprintf("deploy failed: %v", e))
	}
	return ok(result)
}

// HandleCoolifyGetLogs gets service logs.
func HandleCoolifyGetLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serviceUUID, _ := getString(args, "service_uuid", "service_id")
	if serviceUUID == "" {
		return err("service_uuid is required")
	}
	
	path := fmt.Sprintf("/services/%s/logs", serviceUUID)
	if tail := getInt(args, "tail"); tail > 0 {
		path += fmt.Sprintf("?tail=%d", tail)
	}

	result, e := coolifyRequest(ctx, "GET", path, nil)
	if e != nil {
		return err(fmt.Sprintf("get logs failed: %v", e))
	}
	return ok(result)
}

// HandleCoolifyListDatabases lists database resources.
func HandleCoolifyListDatabases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectUUID, _ := getString(args, "project_uuid", "project_id")
	path := "/databases"
	if projectUUID != "" {
		path = fmt.Sprintf("/projects/%s/databases", projectUUID)
	}

	result, e := coolifyRequest(ctx, "GET", path, nil)
	if e != nil {
		return err(fmt.Sprintf("list databases failed: %v", e))
	}
	return ok(result)
}
