package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func HandleInit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dbType, _ :=getString(args, "database_type")
	connStr, _ :=getString(args, "connection_string")
	hostMode, _ :=getString(args, "host_mode")

	if dbType == "" || connStr == "" || hostMode == "" {
		return err("missing required parameters: database_type, connection_string, or host_mode")
}

	cmd := exec.Command("dab", "init",
		"--database-type", dbType,
		"--connection-string", connStr,
		"--host-mode", hostMode)

	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("init failed: %s", string(output)))
}

	return ok("Data API builder initialized successfully")
}

func HandleAddEntity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	entityName, _ :=getString(args, "entity_name")
	source, _ :=getString(args, "source")
	permissions, _ :=getString(args, "permissions")

	if entityName == "" || source == "" || permissions == "" {
		return err("missing required parameters: entity_name, source, or permissions")
}

	cmd := exec.Command("dab", "add", entityName,
		"--source", source,
		"--permissions", permissions)

	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("add entity failed: %s", string(output)))
}

	return ok("Entity added successfully")
}

func HandleStart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configPath, _ :=getString(args, "config_path")
	if configPath == "" {
		configPath = "dab-config.json"
	}

	cmd := exec.Command("dab", "start", "--config", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	startErr := cmd.Start()
	if startErr != nil {
		return err(fmt.Sprintf("failed to start: %v", startErr))
}

	go func() {
		<-ctx.Done()
		cmd.Process.Kill()
	}()

	return ok("Data API builder started successfully")
}

func HandleValidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configPath, _ :=getString(args, "config_path")
	if configPath == "" {
		configPath = "dab-config.json"
	}

	cmd := exec.Command("dab", "validate", "--config", configPath)
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("validation failed: %s", string(output)))
}

	return ok("Configuration is valid")
}

func HandleGenerateConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dbType, _ :=getString(args, "database_type")
	if dbType == "" {
		return err("missing required parameter: database_type")
}

	// Build the project to generate config file
	cmd := exec.Command("dotnet", "build", "-p:generateConfigFileForDbType="+dbType)
	cmd.Dir = filepath.Join("src")
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("config generation failed: %s", string(output)))
}

	configPath := filepath.Join("src", "Service", "dab-config."+dbType+".json")
	configContent, readErr := os.ReadFile(configPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read generated config: %v", readErr))
}

	return ok(string(configContent))
}

func HandleHealthCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiUrl, _ :=getString(args, "api_url")
	if apiUrl == "" {
		apiUrl = "http://localhost:5000/health"
	}

	client := http.DefaultClient
	resp, reqErr := client.Get(apiUrl)
	if reqErr != nil {
		return err(fmt.Sprintf("health check failed: %v", reqErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("health check returned status: %s", resp.Status))
}

	var result map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse health response: %v", parseErr))
}

	status := result["status"]
	if status == nil || status != "healthy" {
		return err("health check returned unhealthy status")
}

	return ok("Data API builder is healthy")
}