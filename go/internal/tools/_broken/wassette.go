package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HandleCreateWorkflow creates a new agentic workflow
func HandleCreateWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflowName, _ :=getString(args, "workflow_name")
	description, _ :=getString(args, "description")
	trigger, _ :=getString(args, "trigger")
	schedule, _ :=getString(args, "schedule")
	tools, _ :=getString(args, "tools")
	toolsets, _ :=getString(args, "toolsets")
	roles, _ :=getString(args, "roles")
	network, _ :=getString(args, "network")
	prompt, _ :=getString(args, "prompt")

	if workflowName == "" {
		return err("workflow_name is required")
}

	// Generate workflow ID from name
	workflowID := toKebabCase(workflowName)
	workflowPath := filepath.Join(".github", "workflows", workflowID+".md")

	// Check if file exists and append suffix if needed
	if _, statErr := os.Stat(workflowPath); statErr == nil {
		timestamp := time.Now().Unix()
		workflowID = fmt.Sprintf("%s-%d", workflowID, timestamp)
		workflowPath = filepath.Join(".github", "workflows", workflowID+".md")

	// Build frontmatter
	frontmatter := buildFrontmatter(workflowName, description, trigger, schedule, tools, toolsets, roles, network)

	// Build workflow content
	content := frontmatter + "\n\n" + prompt

	// Ensure directory exists
	if mkdirErr := os.MkdirAll(filepath.Dir(workflowPath), 0755); mkdirErr != nil {
		return err(fmt.Sprintf("failed to create directory: %v", mkdirErr))
}

	// Write workflow file
	if writeErr := os.WriteFile(workflowPath, []byte(content), 0644); writeErr != nil {
		return err(fmt.Sprintf("failed to write workflow file: %v", writeErr))
}

	// Compile the workflow
	compileOut, compileErr := exec.Command("gh", "aw", "compile", workflowID).CombinedOutput()
	if compileErr != nil {
		return ok(fmt.Sprintf("Created workflow at %s but compilation failed:\n%s\n%s", workflowPath, compileOut, compileErr))
}

	return ok(fmt.Sprintf("Successfully created and compiled workflow '%s' at %s", workflowName, workflowPath))
}

}

// HandleUpdateWorkflow updates an existing agentic workflow
func HandleUpdateWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflowID, _ :=getString(args, "workflow_id")
	field, _ :=getString(args, "field")
	value, _ :=getString(args, "value")

	if workflowID == "" {
		return err("workflow_id is required")
}

	workflowPath := findWorkflowPath(workflowID)
	if workflowPath == "" {
		return err(fmt.Sprintf("workflow '%s' not found", workflowID))
}

	content, readErr := os.ReadFile(workflowPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read workflow file: %v", readErr))
}

	var updatedContent string
	switch field {
	case "description":
		updatedContent = updateFrontmatterField(string(content), "description", value)
	case "prompt", "instructions":
		updatedContent = updatePromptContent(string(content), value)
	default:
		return err(fmt.Sprintf("unknown field '%s'. Supported fields: description, prompt", field))
}

	if writeErr := os.WriteFile(workflowPath, []byte(updatedContent), 0644); writeErr != nil {
		return err(fmt.Sprintf("failed to update workflow file: %v", writeErr))
}

	// Recompile
	compileErr := exec.Command("gh", "aw", "compile", workflowID).Run()
	if compileErr != nil {
		return ok(fmt.Sprintf("Updated %s but recompilation failed: %v", workflowPath, compileErr))
}

	return ok(fmt.Sprintf("Successfully updated workflow '%s' field '%s'", workflowID, field))
}

// HandleDebugWorkflow analyzes and debugs workflow issues
func HandleDebugWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflowID, _ :=getString(args, "workflow_id")
	runID, _ :=getString(args, "run_id")
	action, _ :=getString(args, "action")

	if workflowID == "" && runID == "" {
		return err("either workflow_id or run_id is required")
}

	switch action {
	case "audit":
		return auditWorkflow(workflowID)
}
	case "logs":
		return getWorkflowLogs(workflowID, runID)
}
	case "validate":
		return validateWorkflow(workflowID)
	default:
		// Default to audit
		return auditWorkflow(workflowID)

}

func auditWorkflow(workflowID string) (ToolResponse, error) {
	workflowPath := findWorkflowPath(workflowID)
	if workflowPath == "" {
		return err(fmt.Sprintf("workflow '%s' not found", workflowID))
}

	content, readErr := os.ReadFile(workflowPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read workflow: %v", readErr))
}

	var issues []string
	var suggestions []string

	// Check frontmatter
	if !bytes.Contains(content, []byte("---")) {
		issues = append(issues, "Missing frontmatter delimiters")

	// Check for required fields
	if !bytes.Contains(content, []byte("description:")) {
		issues = append(issues, "Missing description in frontmatter")
		suggestions = append(suggestions, "Add a description field to the frontmatter")

	// Check for tools
	if !bytes.Contains(content, []byte("tools:")) && !bytes.Contains(content, []byte("toolsets:")) {
		issues = append(issues, "No tools or toolsets defined")
		suggestions = append(suggestions, "Add tools or toolsets to specify available capabilities")

	// Check for trigger
	if !bytes.Contains(content, []byte("on:")) && !bytes.Contains(content, []byte("trigger:")) {
		issues = append(issues, "No trigger defined")
		suggestions = append(suggestions, "Add a trigger (issues, pull_requests, schedule, workflow_dispatch)")

	// Check for safe outputs if GitHub write operations
	if bytes.Contains(content, []byte("github")) && !bytes.Contains(content, []byte("safe-outputs")) {
		suggestions = append(suggestions, "Consider using safe-outputs for GitHub write operations")

	// Check prompt length
	lines := strings.Split(string(content), "\n")
	promptLines := 0
	inFrontmatter := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter {
				inFrontmatter = true
			} else {
				break
			}
		}
		if inFrontmatter && !strings.Contains(line, "---") {
			promptLines++
		}
	}

	if promptLines < 5 {
		suggestions = append(suggestions, "Prompt seems short; consider adding more detailed instructions")

	result := fmt.Sprintf("Audit results for '%s':\n", workflowID)
	if len(issues) > 0 {
		result += "\nIssues found:\n"
		for _, issue := range issues {
			result += fmt.Sprintf("  - %s\n", issue)

	}
	if len(suggestions) > 0 {
		result += "\nSuggestions:\n"
		for _, suggestion := range suggestions {
			result += fmt.Sprintf("  - %s\n", suggestion)

	}
	if len(issues) == 0 && len(suggestions) == 0 {
		result += "\nNo issues found. Workflow appears well-structured."
	}

	return ok(result)
}

}
}
}
}
}
}
}

func getWorkflowLogs(workflowID, runID string) (ToolResponse, error) {
	var cmd *exec.Cmd
	if runID != "" {
		cmd = exec.Command("gh", "aw", "audit", runID)
	} else {
		cmd = exec.Command("gh", "aw", "logs", workflowID)

	output, logErr := cmd.CombinedOutput()
	if logErr != nil {
		return err(fmt.Sprintf("failed to get logs: %v\n%s", logErr, output))
}

	return ok(fmt.Sprintf("Workflow logs for '%s':\n%s", workflowID, output))
}

}

func validateWorkflow(workflowID string) (ToolResponse, error) {
	cmd := exec.Command("gh", "aw", "compile", "--strict", workflowID)
	output, compileErr := cmd.CombinedOutput()

	result := fmt.Sprintf("Validation results for '%s':\n%s", workflowID, output)
	if compileErr != nil {
		return ok(result + "\nValidation failed")
}

	return ok(result + "\nValidation passed")
}

// HandleUpgradeWorkflows upgrades workflows to new gh-aw versions
func HandleUpgradeWorkflows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflowID, _ :=getString(args, "workflow_id")
	apply, _ :=getBool(args, "apply")

	cmdArgs := []string{"aw", "fix"}
	if apply {
		cmdArgs = append(cmdArgs, "--write")

	if workflowID != "" {
		cmdArgs = append(cmdArgs, workflowID)

	cmd := exec.Command("gh", cmdArgs[0], cmdArgs[1:]...)
	output, upgradeErr := cmd.CombinedOutput()

	result := fmt.Sprintf("Upgrade results:\n%s", output)
	if upgradeErr != nil {
		return ok(result + "\nUpgrade completed with errors")
}

	return ok(result)
}

}
}

// HandleCreateSharedComponent creates a reusable workflow component
func HandleCreateSharedComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	componentName, _ :=getString(args, "component_name")
	description, _ :=getString(args, "description")
	tools, _ :=getString(args, "tools")
	mcpServer, _ :=getString(args, "mcp_server")

	if componentName == "" {
		return err("component_name is required")
}

	componentID := toKebabCase(componentName)
	componentPath := filepath.Join(".github", "workflows", "shared", componentID+".md")

	// Check if file exists
	if _, statErr := os.Stat(componentPath); statErr == nil {
		return err(fmt.Sprintf("component '%s' already exists at %s", componentName, componentPath))
}

	// Build component content
	content := buildSharedComponent(componentName, description, tools, mcpServer)

	// Ensure directory exists
	if mkdirErr := os.MkdirAll(filepath.Dir(componentPath), 0755); mkdirErr != nil {
		return err(fmt.Sprintf("failed to create directory: %v", mkdirErr))
}

	// Write component file
	if writeErr := os.WriteFile(componentPath, []byte(content), 0644); writeErr != nil {
		return err(fmt.Sprintf("failed to write component file: %v", writeErr))
}

	return ok(fmt.Sprintf("Successfully created shared component '%s' at %s", componentName, componentPath))
}

// HandleBuildDocs builds documentation using mdBook
func HandleBuildDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	open, _ :=getBool(args, "open")

	docsPath := "docs"

	switch action {
	case "serve":
		cmd := exec.Command("mdbook", "serve", docsPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if open {
			// Try to open in browser
			exec.Command("xdg-open", "http://localhost:3000").Run()

		if serveErr := cmd.Run(); serveErr != nil {
			return err(fmt.Sprintf("failed to serve docs: %v", serveErr))
}

		return ok("Documentation server running at http://localhost:3000")
}

	case "build":
		cmd := exec.Command("mdbook", "build", docsPath)
		output, buildErr := cmd.CombinedOutput()
		if buildErr != nil {
			return err(fmt.Sprintf("failed to build docs: %v\n%s", buildErr, output))
}

		return ok(fmt.Sprintf("Documentation built successfully:\n%s", output))
}

	case "watch":
		cmd := exec.Command("mdbook", "serve", docsPath, "--watch")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if watchErr := cmd.Run(); watchErr != nil {
			return err(fmt.Sprintf("failed to watch docs: %v", watchErr))
}

		return ok("Documentation watch mode active")
}

	default:
		return err(fmt.Sprintf("unknown action '%s'. Supported: serve, build, watch", action))

}

// HandleListWorkflows lists all agentic workflows
func HandleListWorkflows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflowsDir := ".github/workflows"

	entries, readErr := os.ReadDir(workflowsDir)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read workflows directory: %v", readErr))
}

	var workflows []map[string]string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".md") && !strings.HasSuffix(entry.Name(), ".lock.yml") {
			workflowID := strings.TrimSuffix(entry.Name(), ".md")
			workflows = append(workflows, map[string]string{
				"id":   workflowID,
				"file": entry.Name(),
			})

	}

	result, jsonErr := json.Marshal(workflows)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal workflows: %v", jsonErr))
}

	return ok(fmt.Sprintf("Found %d workflows:\n%s", len(workflows), result))
}

// HandleCompileWorkflow compiles a workflow to .lock.yml
func HandleCompileWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflowID, _ :=getString(args, "workflow_id")
	strict, _ :=getBool(args, "strict")
	purge, _ :=getBool(args, "purge")

	if workflowID == "" {
		return err("workflow_id is required")
}

	cmdArgs := []string{"aw", "compile", workflowID}
	if strict {
		cmdArgs = append(cmdArgs, "--strict")

	if purge {
		cmdArgs = append(cmdArgs, "--purge")

	cmd := exec.Command("gh", cmdArgs...)
	output, compileErr := cmd.CombinedOutput()

	result := fmt.Sprintf("Compilation results for '%s':\n%s", workflowID, output)
	if compileErr != nil {
		return ok(result + "\nCompilation failed")
}

	return ok(result + "\nCompilation successful")
}

}
}

// Helper functions

func buildFrontmatter(name, desc, trigger, schedule, tools, toolsets, roles, network string) string {
	var fm []string
	fm = append(fm, "---")
	fm = append(fm, fmt.Sprintf("name: %s", name))
	if desc != "" {
		fm = append(fm, fmt.Sprintf("description: %s", desc))

	if trigger != "" {
		fm = append(fm, fmt.Sprintf("on: %s", trigger))

	if schedule != "" {
		fm = append(fm, fmt.Sprintf("schedule: %s", schedule))

	if tools != "" {
		fm = append(fm, fmt.Sprintf("tools: [%s]", tools))

	if toolsets != "" {
		fm = append(fm, fmt.Sprintf("toolsets: [%s]", toolsets))

	if roles != "" {
		fm = append(fm, fmt.Sprintf("roles: [%s]", roles))

	if network != "" {
		fm = append(fm, fmt.Sprintf("network: %s", network))

	fm = append(fm, "---")
	return strings.Join(fm, "\n")
}

}
}
}
}
}
}
}

func buildSharedComponent(name, desc, tools, mcpServer string) string {
	var content []string
	content = append(content, "---")
	content = append(content, fmt.Sprintf("name: %s", name))
	if desc != "" {
		content = append(content, fmt.Sprintf("description: %s", desc))

	content = append(content, "type: shared")
	if tools != "" {
		content = append(content, fmt.Sprintf("tools: [%s]", tools))

	if mcpServer != "" {
		content = append(content, fmt.Sprintf("mcp_server: %s", mcpServer))

	content = append(content, "---")
	content = append(content, "")
	content = append(content, "# Shared Component: "+name)
	if desc != "" {
		content = append(content, "")
		content = append(content, desc)

	return strings.Join(content, "\n")
}

}
}
}
}

func updateFrontmatterField(content, field, value string) string {
}