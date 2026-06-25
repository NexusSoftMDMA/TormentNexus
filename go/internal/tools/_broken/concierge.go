package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// HandleInitProject initializes a new Concierge project
func HandleInitProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("project name is required")
}

	// Create project directory
	projectDir := name
	if e := os.MkdirAll(projectDir, 0755); e != nil {
		return err(fmt.Sprintf("failed to create project directory: %v", e))
}

	// Create main.py
	mainContent := `"""Concierge MCP Server - """ + name + `"""")

from concierge import Concierge

app = Concierge("` + name + `")

# Define your tools here
@app.tool()
def example_tool(query: str) -> dict:
    """Example tool for the workflow."""
    return {"result": f"Processed: {query}"}
}

# Define workflow stages
app.stages = {
    "start": ["example_tool"],
}

# Define stage transitions
app.transitions = {
    "start": [],
}

if __name__ == "__main__":
    app.run()
`

	if e := os.WriteFile(filepath.Join(projectDir, "main.py"), []byte(mainContent), 0644); e != nil {
		return err(fmt.Sprintf("failed to create main.py: %v", e))
}

	// Create requirements.txt
	reqContent := "concierge-sdk>=0.15.0\n"
	if e := os.WriteFile(filepath.Join(projectDir, "requirements.txt"), []byte(reqContent), 0644); e != nil {
		return err(fmt.Sprintf("failed to create requirements.txt: %v", e))
}

	// Create assets directory
	assetsDir := filepath.Join(projectDir, "assets")
	if e := os.MkdirAll(assetsDir, 0755); e != nil {
		return err(fmt.Sprintf("failed to create assets directory: %v", e))
}

	result := map[string]interface{}{
		"project":   name,
		"directory": projectDir,
		"files": []map[string]string{
			{"path": "main.py", "description": "Main server file"},
			{"path": "requirements.txt", "description": "Python dependencies"},
			{"path": "assets/", "description": "Static assets directory"},
		},
		"next_steps": []string{
			"cd " + name,
			"pip install -r requirements.txt",
			"python main.py",
		},
	}

	return ok(result)

// HandleAddTool adds a tool to a Concierge project
func HandleAddTool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ :=getString(args, "project_path")
	toolName, _ :=getString(args, "tool_name")
	description, _ :=getString(args, "description")
	params, _ :=getString(args, "parameters") // JSON string of parameters

	if projectPath == "" || toolName == "" {
		return err("project_path and tool_name are required")
}

	mainPath := filepath.Join(projectPath, "main.py")
	content, readErr := os.ReadFile(mainPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read main.py: %v", readErr))
}

	// Parse parameters if provided
	var paramMap map[string]string
	if params != "" {
		if parseErr := json.Unmarshal([]byte(params), &paramMap); parseErr != nil {
			return err(fmt.Sprintf("invalid parameters JSON: %v", parseErr))

	}

	// Build tool function
	var paramList []string
	for k, v := range paramMap {
		paramList = append(paramList, fmt.Sprintf("%s: %s", k, v))

	paramStr := strings.Join(paramList, ", ")

	newTool := fmt.Sprintf(`

@app.tool()
def %s(%s) -> dict:
    """%s"""
    return {"status": "success", "tool": "%s"}
}
`, toolName, paramStr, description, toolName)

	// Append to main.py
	newContent := string(content) + newTool
	if writeErr := os.WriteFile(mainPath, []byte(newContent), 0644); writeErr != nil {
		return err(fmt.Sprintf("failed to write main.py: %v", writeErr))
}

	return ok(map[string]interface{}{
}
		"tool":      toolName,
		"project":   projectPath,
		"file":      mainPath,
		"added":     true,
		"signature": fmt.Sprintf("def %s(%s)", toolName, paramStr),
	})

}

// HandleSetStages defines workflow stages for a Concierge project
func HandleSetStages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ :=getString(args, "project_path")
	stagesJSON, _ :=getString(args, "stages") // JSON object of stage -> tools

	if projectPath == "" || stagesJSON == "" {
		return err("project_path and stages are required")
}

	mainPath := filepath.Join(projectPath, "main.py")
	content, readErr := os.ReadFile(mainPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read main.py: %v", readErr))
}

	// Validate and format stages JSON
	var stages map[string][]string
	if parseErr := json.Unmarshal([]byte(stagesJSON), &stages); parseErr != nil {
		return err(fmt.Sprintf("invalid stages JSON: %v", parseErr))
}

	// Format as Python dict
	stagesFormatted, formatErr := formatStagesAsPython(stages)
	if formatErr != nil {
		return err(fmt.Sprintf("failed to format stages: %v", formatErr))
}

	// Check if stages already exist and replace
	contentStr := string(content)
	if strings.Contains(contentStr, "app.stages =") {
		// Replace existing stages
		startIdx := strings.Index(contentStr, "app.stages =")
		endIdx := strings.Index(contentStr[startIdx:], "\n\n")
		if endIdx == -1 {
			endIdx = len(contentStr)
		} else {
			endIdx += startIdx
		}
		contentStr = contentStr[:startIdx] + "app.stages = " + stagesFormatted + contentStr[endIdx:]
	} else {
		// Add stages before app.run()
		if runIdx := strings.Index(contentStr, "app.run()"); runIdx != -1 {
			contentStr = contentStr[:runIdx] + "app.stages = " + stagesFormatted + "\n\n" + contentStr[runIdx:]
		} else {
			contentStr += "\napp.stages = " + stagesFormatted + "\n"
		}
	}

	if writeErr := os.WriteFile(mainPath, []byte(contentStr), 0644); writeErr != nil {
		return err(fmt.Sprintf("failed to write main.py: %v", writeErr))
}

	return ok(map[string]interface{}{
}
		"project": projectPath,
		"stages":  stages,
		"updated": true,
	})

// HandleSetTransitions defines stage transitions for a Concierge project
func HandleSetTransitions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ :=getString(args, "project_path")
	transitionsJSON, _ :=getString(args, "transitions") // JSON object of stage -> next_stages

	if projectPath == "" || transitionsJSON == "" {
		return err("project_path and transitions are required")
}

	mainPath := filepath.Join(projectPath, "main.py")
	content, readErr := os.ReadFile(mainPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read main.py: %v", readErr))
}

	// Validate and format transitions JSON
	var transitions map[string][]string
	if parseErr := json.Unmarshal([]byte(transitionsJSON), &transitions); parseErr != nil {
		return err(fmt.Sprintf("invalid transitions JSON: %v", parseErr))
}

	// Format as Python dict
	transFormatted, formatErr := formatTransitionsAsPython(transitions)
	if formatErr != nil {
		return err(fmt.Sprintf("failed to format transitions: %v", formatErr))
}

	// Check if transitions already exist and replace
	contentStr := string(content)
	if strings.Contains(contentStr, "app.transitions =") {
		// Replace existing transitions
		startIdx := strings.Index(contentStr, "app.transitions =")
		endIdx := strings.Index(contentStr[startIdx:], "\n\n")
		if endIdx == -1 {
			endIdx = strings.Index(contentStr[startIdx:], "\n")
			if endIdx == -1 {
				endIdx = len(contentStr)
			} else {
				endIdx += startIdx
			}
		} else {
			endIdx += startIdx
		}
		contentStr = contentStr[:startIdx] + "app.transitions = " + transFormatted + contentStr[endIdx:]
	} else {
		// Add transitions before app.run()
		if runIdx := strings.Index(contentStr, "app.run()"); runIdx != -1 {
			contentStr = contentStr[:runIdx] + "app.transitions = " + transFormatted + "\n\n" + contentStr[runIdx:]
		} else {
			contentStr += "\napp.transitions = " + transFormatted + "\n"
		}
	}

	if writeErr := os.WriteFile(mainPath, []byte(contentStr), 0644); writeErr != nil {
		return err(fmt.Sprintf("failed to write main.py: %v", writeErr))
}

	return ok(map[string]interface{}{
}
		"project":     projectPath,
		"transitions": transitions,
		"updated":     true,
	})

// HandleSetState sets workflow state for a session
func HandleSetState(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	sessionID, _ :=getString(args, "session_id")

	if key == "" {
		return err("key is required")
}

	// Create state directory
	stateDir := filepath.Join(os.TempDir(), "concierge_state", sessionID)
	if mkdirErr := os.MkdirAll(stateDir, 0755); mkdirErr != nil {
		return err(fmt.Sprintf("failed to create state directory: %v", mkdirErr))
}

	// Store state as JSON file
	stateFile := filepath.Join(stateDir, key+".json")
	stateData := map[string]interface{}{
		"key":       key,
		"value":     value,
		"session":   sessionID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	jsonData, marshalErr := json.MarshalIndent(stateData, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal state: %v", marshalErr))
}

	if writeErr := os.WriteFile(stateFile, jsonData, 0644); writeErr != nil {
		return err(fmt.Sprintf("failed to write state file: %v", writeErr))
}

	return ok(map[string]interface{}{
}
		"key":        key,
		"session_id": sessionID,
		"stored":     true,
		"path":       stateFile,
	})

// HandleGetState retrieves workflow state for a session
func HandleGetState(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	sessionID, _ :=getString(args, "session_id")

	if key == "" {
		return err("key is required")
}

	// Read state file
	stateDir := filepath.Join(os.TempDir(), "concierge_state", sessionID)
	stateFile := filepath.Join(stateDir, key+".json")

	content, readErr := os.ReadFile(stateFile)
	if readErr != nil {
		return err(fmt.Sprintf("state not found for key '%s' in session '%s'", key, sessionID))
}

	var stateData map[string]interface{}
	if parseErr := json.Unmarshal(content, &stateData); parseErr != nil {
		return err(fmt.Sprintf("failed to parse state: %v", parseErr))
}

	return ok(stateData)
}

// HandleRunServer runs a Concierge server
func HandleRunServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ :=getString(args, "project_path")
	transport, _ :=getString(args, "transport")

	if projectPath == "" {
		return err("project_path is required")
}

	if transport == "" {
		transport = "stdio"
	}

	mainPath := filepath.Join(projectPath, "main.py")
	if _, statErr := os.Stat(mainPath); statErr != nil {
		return err(fmt.Sprintf("main.py not found in %s", projectPath))
}

	// Check if concierge is installed
	checkCmd := exec.Command("python", "-c", "import concierge")
	if checkErr := checkCmd.Run(); checkErr != nil {
		return err("concierge-sdk is not installed. Run: pip install concierge-sdk")
}

	// Build command based on transport
	var cmd *exec.Cmd
	switch transport {
	case "http":
		cmd = exec.Command("python", "-c", fmt.Sprintf(`
from concierge import Concierge
import json

# Load and run the project
exec(open('%s').read())
http_app = app.streamable_http_app()
import uvicorn
uvicorn.run(http_app, host="0.0.0.0", port=8000)
`, filepath.Abs(mainPath)))
	case "sse":
		cmd = exec.Command("python", mainPath)
	default:
		cmd = exec.Command("python", mainPath)

	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if startErr := cmd.Start(); startErr != nil {
		return err(fmt.Sprintf("failed to start server: %v", startErr))
}

	return ok(map[string]interface{}{
}
		"project":   projectPath,
		"transport": transport,
		"pid":       cmd.Process.Pid,
		"started":   true,
		"message":   fmt.Sprintf("Server started with PID %d", cmd.Process.Pid),
	})

}

// Helper function to format stages as Python dict
func formatStagesAsPython(stages map[string][]string) (string, error) {
	var lines []string
	for stage, tools := range stages {
		toolsList := make([]string, len(tools))
		for i, t := range tools {
			toolsList[i] = fmt.Sprintf(`"%s"`, t)

		lines = append(lines, fmt.Sprintf(`    "%s": [%s]`, stage, strings.Join(toolsList, ", ")))

	return "{\n" + strings.Join(lines, ",\n") + "\n}", nil
}

}
}

// Helper function to format transitions as Python dict
func formatTransitionsAsPython(transitions map[string][]string) (string, error) {
	var lines []string
	for stage, nextStages := range transitions {
		if len(nextStages) == 0 {
			lines = append(lines, fmt.Sprintf(`    "%s": []`, stage))
		} else {
			stagesList := make([]string, len(nextStages))
			for i, s := range nextStages {
				stagesList[i] = fmt.Sprintf(`"%s"`, s)

			lines = append(lines, fmt.Sprintf(`    "%s": [%s]`, stage, strings.Join(stagesList, ", ")))

	}
	return "{\n" + strings.Join(lines, ",\n") + "\n}", nil
}
}
}