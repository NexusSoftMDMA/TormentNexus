//go:build ignore
// +build ignore

package tools

/**
 * @file docker_mcp.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Docker MCP — container management.
 * Replaces: Docker MCP server
 *
 * Provides Docker container management through the Docker Engine API.
 * Configurable via DOCKER_HOST (default unix:///var/run/docker.sock).
 *
 * Tools:
 *  - docker_list_containers — list containers
 *  - docker_list_images — list images
 *  - docker_inspect — inspect a container
 *  - docker_logs — get container logs
 *  - docker_stats — container resource stats
 *  - docker_exec — run a command in a container
 */

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

func dockerClient() *http.Client {
	return &http.Client{Timeout: 60 * time.Second}
}

func dockerURL() string {
	if u := os.Getenv("DOCKER_HOST"); u != "" {
		return u
	}
	return "http://localhost/v1.45"
}

func dockerRequest(ctx context.Context, method, path string, body io.Reader) (string, error) {
	req, e := http.NewRequestWithContext(ctx, method, dockerURL()+path, body)
	if e != nil {
		return "", fmt.Errorf("request error: %v", e)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, e := dockerClient().Do(req)
	if e != nil {
		return "", fmt.Errorf("docker API error: %v", e)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("Docker API error (%d): %s", resp.StatusCode, string(data))
	}
	return string(data), nil
}

// HandleDockerListContainers lists Docker containers.
func HandleDockerListContainers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	all := "false"
	if getBool(args, "all", "show_all") {
		all = "true"
	}
	limit := ""
	if l := getInt(args, "limit"); l > 0 {
		limit = fmt.Sprintf("&limit=%d", l)
	}
	result, e := dockerRequest(ctx, "GET", fmt.Sprintf("/containers/json?all=%s%s", all, limit), nil)
	if e != nil {
		return err(fmt.Sprintf("list containers failed: %v", e))
	}
	return ok(result)
}

// HandleDockerListImages lists Docker images.
func HandleDockerListImages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	all := "false"
	if getBool(args, "all") {
		all = "true"
	}
	result, e := dockerRequest(ctx, "GET", fmt.Sprintf("/images/json?all=%s", all), nil)
	if e != nil {
		return err(fmt.Sprintf("list images failed: %v", e))
	}
	return ok(result)
}

// HandleDockerInspect inspects a container.
func HandleDockerInspect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ := getString(args, "id", "container", "name")
	if id == "" {
		return err("container id or name is required")
	}
	result, e := dockerRequest(ctx, "GET", fmt.Sprintf("/containers/%s/json", id), nil)
	if e != nil {
		return err(fmt.Sprintf("inspect failed: %v", e))
	}
	return ok(result)
}

// HandleDockerLogs gets container logs.
func HandleDockerLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ := getString(args, "id", "container", "name")
	if id == "" {
		return err("container id or name is required")
	}
	tail := "50"
	if l := getInt(args, "tail"); l > 0 {
		tail = fmt.Sprintf("%d", l)
	}
	result, e := dockerRequest(ctx, "GET", fmt.Sprintf("/containers/%s/logs?stdout=true&stderr=true&tail=%s", id, tail), nil)
	if e != nil {
		return err(fmt.Sprintf("logs failed: %v", e))
	}
	return ok(result)
}

// HandleDockerStats gets container resource statistics.
func HandleDockerStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ := getString(args, "id", "container", "name")
	if id == "" {
		return err("container id or name is required")
	}
	result, e := dockerRequest(ctx, "GET", fmt.Sprintf("/containers/%s/stats?stream=false", id), nil)
	if e != nil {
		return err(fmt.Sprintf("stats failed: %v", e))
	}
	return ok(result)
}

// HandleDockerExec runs a command in a container.
func HandleDockerExec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ := getString(args, "id", "container", "name")
	if id == "" {
		return err("container id or name is required")
	}
	cmd, _ := getString(args, "cmd", "command", "exec")
	if cmd == "" {
		return err("command is required")
	}

	// Create exec instance
	execBody := map[string]interface{}{
		"Cmd":          strings.Fields(cmd),
		"AttachStdout": true,
		"AttachStderr": true,
	}
	b, _ := json.Marshal(execBody)
	execResult, e := dockerRequest(ctx, "POST", fmt.Sprintf("/containers/%s/exec", id), strings.NewReader(string(b)))
	if e != nil {
		return err(fmt.Sprintf("exec create failed: %v", e))
	}

	// Parse exec ID
	var execResp struct {
		ID string `json:"Id"`
	}
	if e := json.Unmarshal([]byte(execResult), &execResp); e != nil || execResp.ID == "" {
		return err("failed to create exec instance")
	}

	// Start exec
	startBody := map[string]interface{}{
		"Detach": false,
		"Tty":    false,
	}
	sb, _ := json.Marshal(startBody)
	startResult, e := dockerRequest(ctx, "POST", fmt.Sprintf("/exec/%s/start", execResp.ID), strings.NewReader(string(sb)))
	if e != nil {
		return err(fmt.Sprintf("exec start failed: %v", e))
	}

	return ok(startResult)
}
