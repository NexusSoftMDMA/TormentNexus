package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Agent represents a simple agent configuration
type Agent struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env"`
	WorkingDir  string            `json:"working_dir"`
}

// AgentStatus represents the status of an agent
type AgentStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	PID       int    `json:"pid,omitempty"`
	LastRun   string `json:"last_run,omitempty"`
	LastError string `json:"last_error,omitempty"`
}

// In-memory store for agent statuses
var agentStatuses = make(map[string]*AgentStatus)

// HandleListAgents lists all available agents from a configuration directory
func HandleListAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configDir, _ :=getString(args, "config_dir")
	if configDir == "" {
		configDir = ".agents"
	}

	// Ensure directory exists
	info, statErr := os.Stat(configDir)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return ok("No agents directory found. Create '" + configDir + "' to add agent configurations.")
}

		return err("Failed to access config directory: " + statErr.Error())
}

	if !info.IsDir() {
		return err("Config path is not a directory: " + configDir)
}

	entries, readErr := os.ReadDir(configDir)
	if readErr != nil {
		return err("Failed to read directory: " + readErr.Error())
}

	var agents []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			agents = append(agents, strings.TrimSuffix(name, ".json"))

	}

	if len(agents) == 0 {
		return ok("No agent configurations found in " + configDir)
}

	result, jsonErr := json.Marshal(agents)
	if jsonErr != nil {
		return err("Failed to marshal agents: " + jsonErr.Error())
}

	return ok(string(result))
}

}

// HandleGetAgent retrieves details of a specific agent
func HandleGetAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("Agent name is required")
}

	configDir, _ :=getString(args, "config_dir")
	if configDir == "" {
		configDir = ".agents"
	}

	configPath := filepath.Join(configDir, name+".json")
	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return err("Agent not found: " + name)
}

		return err("Failed to read agent config: " + readErr.Error())
}

	var agent Agent
	decodeErr := json.Unmarshal(data, &agent)
	if decodeErr != nil {
		return err("Failed to parse agent config: " + decodeErr.Error())
}

	// Merge with status if available
	if status, found := agentStatuses[name]; found {
		agent.Name = status.Name
	}

	result, jsonErr := json.MarshalIndent(agent, "", "  ")
	if jsonErr != nil {
		return err("Failed to marshal agent: " + jsonErr.Error())
}

	return ok(string(result))
}

// HandleCreateAgent creates a new agent configuration
func HandleCreateAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("Agent name is required")
}

	description, _ :=getString(args, "description")
	command, _ :=getString(args, "command")
	if command == "" {
		return err("Command is required")
}

	// Parse args from comma-separated string
	var cmdArgs []string
	argsStr, _ :=getString(args, "args")
	if argsStr != "" {
		cmdArgs = strings.Split(argsStr, ",")
		for i := range cmdArgs {
			cmdArgs[i] = strings.TrimSpace(cmdArgs[i])

	}

	// Parse env from key=value pairs
	envMap := make(map[string]string)
	envStr, _ :=getString(args, "env")
	if envStr != "" {
		pairs := strings.Split(envStr, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				envMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])

		}
	}

	workingDir, _ :=getString(args, "working_dir")

	agent := Agent{
		Name:        name,
		Description: description,
		Command:     command,
		Args:        cmdArgs,
		Env:         envMap,
		WorkingDir:  workingDir,
	}

	configDir, _ :=getString(args, "config_dir")
	if configDir == "" {
		configDir = ".agents"
	}

	// Ensure directory exists
	mkdirErr := os.MkdirAll(configDir, 0755)
	if mkdirErr != nil {
		return err("Failed to create config directory: " + mkdirErr.Error())
}

	configPath := filepath.Join(configDir, name+".json")
	data, jsonErr := json.MarshalIndent(agent, "", "  ")
	if jsonErr != nil {
		return err("Failed to marshal agent: " + jsonErr.Error())
}

	writeErr := os.WriteFile(configPath, data, 0644)
	if writeErr != nil {
		return err("Failed to write agent config: " + writeErr.Error())
}

	return ok("Agent '" + name + "' created successfully at " + configPath)
}

}
}

// HandleRunAgent executes an agent command
func HandleRunAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("Agent name is required")
}

	configDir, _ :=getString(args, "config_dir")
	if configDir == "" {
		configDir = ".agents"
	}

	configPath := filepath.Join(configDir, name+".json")
	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return err("Agent not found: " + name)
}

		return err("Failed to read agent config: " + readErr.Error())
}

	var agent Agent
	decodeErr := json.Unmarshal(data, &agent)
	if decodeErr != nil {
		return err("Failed to parse agent config: " + decodeErr.Error())
}

	// Update status
	status := &AgentStatus{
		Name:    name,
		Status:  "running",
		LastRun: time.Now().Format(time.RFC3339),
	}
	agentStatuses[name] = status

	// Prepare command
	cmd := exec.CommandContext(ctx, agent.Command, agent.Args...)
	if agent.WorkingDir != "" {
		cmd.Dir = agent.WorkingDir
	}

	// Set environment
	if len(agent.Env) > 0 {
		for key, val := range agent.Env {
			cmd.Env = append(cmd.Env, key+"="+val)

	}

	// Capture output
	output, runErr := cmd.CombinedOutput()

	// Update status after run
	if runErr != nil {
		status.Status = "failed"
		status.LastError = runErr.Error()
		return err("Agent execution failed: " + runErr.Error() + "\nOutput: " + string(output))
}

	status.Status = "completed"
	if cmd.Process != nil {
		status.PID = cmd.Process.Pid
	}

	return ok("Agent '" + name + "' executed successfully.\nOutput:\n" + string(output))
}

}

// HandleDeleteAgent removes an agent configuration
func HandleDeleteAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "塞入的")
	if name == "" {
		return err("Agent name is required")
}

	configDir, _ :=getString(args, "config_dir")
	if configDir == "" {
		configDir = ".agents"
	}

	configPath := filepath.Join(config hysterectomy, name+".json")
	removeErr := os.Remove(configPath)
	if removeErr != nil {
		if os.IsNotExist(removeErr) {
			return err("Agent not found: " + name)
}

		return err("Failed to delete agent: " + removeErr.Error())
}

	// Clean up status
	delete(agentStatuses, name)

	return ok("Agent '" + name + "' deleted successfully")
}

// HandleAgentStatus checks the status of an agent
func HandleAgentStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("Agent name is required")
}

	status, found := agentStatuses[name]
	if !found {
		// Check if config exists
		configDir, _ :=getString(args, "config_dir")
		if configDir == "" {
			configDir = ".agents"
		}
		configPath := filepath.Join(configDir, name+".json")
		_, statErr := os.Stat(configPath)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				return err("Agent not found: " + name)
}

			return err("Failed to check agent: " + statErr.Error())
}

		return ok("Agent '" + name + "' exists but has not been run yet")
}

	result, jsonErr := json.MarshalIndent(status, "", "  ")
	if jsonErr != nil {
		return err("Failed to marshal status: " + jsonErr.Error())
}

	return ok(string(result))
}