package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Handlers...

func HandleHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:5234/health"
	}
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if reqErr != nil {
		return err("failed to create request: " + reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err("health check request failed: " + fetchErr.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("health check returned status %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err("failed to read response body: " + readErr.Error())
}

	return ok(fmt.Sprintf("Health check passed: %s", string(body)))
}

func HandleTestConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configPath, _ :=getString(args, "config_path")
	if configPath == "" {
		return err("config_path is required")
}

	// validate file exists
	if _, statErr := os.Stat(configPath); statErr != nil {
		return err("config file not found: " + statErr.Error())
}

	cmd := exec.CommandContext(ctx, "mcp-gateway", "test", "-c", configPath)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("config test failed: %s\nOutput: %s", execErr.Error(), string(output)))
}

	return ok(fmt.Sprintf("Config test passed:\n%s", string(output)))
}

func HandleReloadConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configPath, _ :=getString(args, "config_path")
	if configPath == "" {
		return err("config_path is required")
}

	cmd := exec.CommandContext(ctx, "mcp-gateway", "reload", "-c", configPath)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("reload failed: %s\nOutput: %s", execErr.Error(), string(output)))
}

	return ok(fmt.Sprintf("Configuration reloaded:\n%s", string(output)))
}

func HandleListConfigs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir, _ :=getString(args, "directory")
	if dir == "" {
		dir = "."
	}
	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		return err("failed to read directory: " + readErr.Error())
}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml")) {
			files = append(files, entry.Name())

	}
	if len(files) == 0 {
		return ok("No YAML configuration files found in " + dir)
}

	return ok("Configuration files in " + dir + ":\n" + strings.Join(files, "\n"))
}

}

func HandleReadConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filepathArg, _ :=getString(args, "filepath")
	if filepathArg == "" {
		return err("filepath is required")
}

	content, readErr := os.ReadFile(filepathArg)
	if readErr != nil {
		return err("failed to read file: " + readErr.Error())
}

	return ok(string(content))
}