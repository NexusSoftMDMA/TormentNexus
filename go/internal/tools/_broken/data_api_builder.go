package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// runDabCommand executes the DAB CLI with the given arguments.
func runDabCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "dab", args...)
	output, cmdErr := cmd.CombinedOutput()
	return string(output), cmdErr
}

// HandleInit initializes a new Data API builder configuration file.
func HandleInit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dbType, _ :=getString(args, "database_type")
	connEnvVar, _ :=getString(args, "connection_string_env_var")
	hostMode, _ :=getString(args, "host_mode")

	if hostMode == "" {
		hostMode = "Production"
	}

	// Construct arguments for 'dab init'
	// Example: dab init --database-type mssql --connection-string "@env('MY_CONN_STR')" --host-mode Production
	cmdArgs := []string{
		"init",
		"--database-type", dbType,
		"--connection-string", fmt.Sprintf("@env('%s')", connEnvVar),
		"--host-mode", hostMode,
	}

	output, cmdErr := runDabCommand(ctx, cmdArgs...)
	if cmdErr != nil {
		return err(fmt.Sprintf("failed to initialize DAB: %s\nOutput: %s", cmdErr.Error(), output))
}

	return ok(fmt.Sprintf("Successfully initialized DAB configuration.\n%s", output))
}

// HandleAddEntity adds a new entity to the DAB configuration.
func HandleAddEntity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	entityName, _ :=getString(args, "entity_name")
	source, _ :=getString(args, "source")
	permissions, _ :=getString(args, "permissions")
	configFile, _ :=getString(args, "config_file")

	if configFile == "" {
		configFile = "dab-config.json"
	}

	// Construct arguments for 'dab add'
	// Example: dab add MyEntity --source dbo.MyTable --permissions "anonymous:*"
	cmdArgs := []string{
		"add", entityName,
		"--source", source,
		"--permissions", permissions,
		"--config", configFile,
	}

	output, cmdErr := runDabCommand(ctx, cmdArgs...)
	if cmdErr != nil {
		return err(fmt.Sprintf("failed to add entity '%s': %s\nOutput: %s", entityName, cmdErr.Error(), output))
}

	return ok(fmt.Sprintf("Successfully added entity '%s'.\n%s", entityName, output))
}

// HandleValidate validates the DAB configuration file.
func HandleValidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configFile, _ :=getString(args, "config_file")

	if configFile == "" {
		configFile = "dab-config.json"
	}

	cmdArgs := []string{"validate", "--config", configFile}
	output, cmdErr := runDabCommand(ctx, cmdArgs...)
	if cmdErr != nil {
		return err(fmt.Sprintf("validation failed: %s\nOutput: %s", cmdErr.Error(), output))
}

	return ok(fmt.Sprintf("Configuration is valid.\n%s", output))
}

// HandleStart starts the Data API builder runtime.
// Note: This command is blocking and will run until the process is stopped.
func HandleStart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configFile, _ :=getString(args, "config_file")

	if configFile == "" {
		configFile = "dab-config.json"
	}

	cmdArgs := []string{"start", "--config", configFile}

	// Execute the command. This will block until the server stops or context is cancelled.
	output, cmdErr := runDabCommand(ctx, cmdArgs...)
	if cmdErr != nil {
		return err(fmt.Sprintf("DAB runtime exited with error: %s\nOutput: %s", cmdErr.Error(), output))
}

	return ok(fmt.Sprintf("DAB runtime stopped.\n%s", output))
}

// HandleReadConfig reads and pretty-prints the DAB configuration file.
func HandleReadConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configFile, _ :=getString(args, "config_file")

	if configFile == "" {
		configFile = "dab-config.json"
	}

	absPath, pathErr := filepath.Abs(configFile)
	if pathErr != nil {
		return err(fmt.Sprintf("failed to resolve path: %s", pathErr.Error()))
}

	data, readErr := os.ReadFile(absPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read config file: %s", readErr.Error()))
}

	// Attempt to pretty-print JSON
	var prettyJSON interface{}
	jsonErr := json.Unmarshal(data, &prettyJSON)
	if jsonErr != nil {
		// If it's not valid JSON, return raw content
		return ok(string(data))
}

	formatted, marshalErr := json.MarshalIndent(prettyJSON, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to format JSON: %s", marshalErr.Error()))
}

	return ok(string(formatted))
}