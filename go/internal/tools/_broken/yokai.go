package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func HandleVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version, _ :=getString(args, "version")
	if version == "" {
		return err("version parameter is required")
}

	return ok(fmt.Sprintf("Yokai version: %s", version))
}

func HandleConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configPath, _ :=getString(args, "path")
	if configPath == "" {
		return err("path parameter is required")
}

	fileInfo, fileErr := os.Stat(configPath)
	if fileErr != nil {
		return err(fmt.Sprintf("failed to read config file: %v", fileErr))
}

	if fileInfo.IsDir() {
		return err("path must point to a file, not a directory")
}

	configContent, readErr := os.ReadFile(configPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read config file: %v", readErr))
}

	return ok(string(configContent))
}

func HandleHealthCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	healthURL, _ :=getString(args, "url")
	if healthURL == "" {
		return err("url parameter is required")
}

	client := http.Client{Timeout: 30 * time.Second}
	resp, reqErr := client.Get(healthURL)
	if reqErr != nil {
		return err(fmt.Sprintf("health check request failed: %v", reqErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("health check failed with status: %s", resp.Status))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	return ok(string(body))
}

func HandleExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command parameter is required")
}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("command execution failed: %v", execErr))
}

	return ok(string(output))
}

func HandleValidateConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configPath, _ :=getString(args, "path")
	if configPath == "" {
		return err("path parameter is required")
}

	configContent, readErr := os.ReadFile(configPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read config file: %v", readErr))
}

	var configData map[string]interface{}
	parseErr := json.Unmarshal(configContent, &configData)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse config file: %v", parseErr))
}

	requiredKeys := []string{"app", "config"}
	for _, key := range requiredKeys {
		if _, exists := configData[key]; !exists {
			return err(fmt.Sprintf("missing required key in config: %s", key))

	}

	return ok("config validation successful")
}

}

func HandleListFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	directory, _ :=getString(args, "directory")
	if directory == "" {
		return err("directory parameter is required")
}

	files, listErr := os.ReadDir(directory)
	if listErr != nil {
		return err(fmt.Sprintf("failed to list files: %v", listErr))
}

	var fileList []string
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, file.Name())
	b	}
	}

	return ok(strings.Join(fileList, "\n"))
}