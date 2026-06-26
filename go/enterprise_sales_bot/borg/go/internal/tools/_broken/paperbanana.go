package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// HandleGenerateDiagram calls the PaperBanana CLI to generate a diagram.
// Expected args:
//   - source_context (string): methodology text.
//   - caption (string): figure caption.
//   - iterations (int, optional): number of refinement iterations (default 3).
func HandleGenerateDiagram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ :=getString(args, "source_context")
	if source == "" {
		return err("source_context is required")
}

	caption, _ :=getString(args, "caption")
	if caption == "" {
		return err("caption is required")
}

	iterations, _ :=getInt(args, "iterations")
	if iterations == 0 {
		iterations = 3
	}

	// Write source context to a temporary file
	tmpFile, tmpErr := ioutil.TempFile("", "paperbanana_source_*.txt")
	if tmpErr != nil {
		return err(fmt.Sprintf("failed to create temp file: %v", tmpErr))
}

	defer os.Remove(tmpFile.Name())

	if _, writeErr := tmpFile.WriteString(source); writeErr != nil {
		return err(fmt.Sprintf("failed to write source context: %v", writeErr))
}

	if closeErr := tmpFile.Close(); closeErr != nil {
		return err(fmt.Sprintf("failed to close temp file: %v", closeErr))
}

	// Build command arguments
	cmdArgs := []string{
		"generate",
		"--input", tmpFile.Name(),
		"--caption", caption,
		"--iterations", strconv.Itoa(iterations),
	}
	cmd := exec.CommandContext(ctx, "paperbanana", cmdArgs...)
	cmd.Dir = filepath.Dir(tmpFile.Name())

	// Execute command
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("paperbanana generate failed: %v, output: %s", execErr, string(output)))
}

	return ok(string(output))
}

// HandleEvaluateDiagram calls the PaperBanana CLI to evaluate a diagram.
// Expected args:
//   - generated_path (string): path to generated image.
//   - reference_path (string): path to reference image.
//   - context (string): methodology text.
//   - caption (string): figure caption.
func HandleEvaluateDiagram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	genPath, _ :=getString(args, "generated_path")
	if genPath == "" {
		return err("generated_path is required")
}

	refPath, _ :=getString(args, "reference_path")
	if refPath == "" {
		return err("reference_path is required")
}

	contextStr, _ :=getString(args, "context")
	if contextStr == "" {
		return err("context is required")
}

	caption, _ :=getString(args, "caption")
	if caption == "" {
		return err("caption is required")
}

	// Write context to a temporary file (CLI expects a file path)
	tmpFile, tmpErr := ioutil.TempFile("", "paperbanana_context_*.txt")
	if tmpErr != nil {
		return err(fmt.Sprintf("failed to create temp file for context: %v", tmpErr))
}

	defer os.Remove(tmpFile.Name())

	if _, writeErr := tmpFile.WriteString(contextStr); writeErr != nil {
		return err(fmt.Sprintf("failed to write context: %v", writeErr))
}

	if closeErr := tmpFile.Close(); closeErr != nil {
		return err(fmt.Sprintf("failed to close context temp file: %v", closeErr))
}

	cmdArgs := []string{
		"evaluate",
		"--generated", genPath,
		"--reference", refPath,
		"--context", tmpFile.Name(),
		"--caption", caption,
	}
	cmd := exec.CommandContext(ctx, "paperbanana", cmdArgs...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("paperbanana evaluate failed: %v, output: %s", execErr, string(output)))
}

	return ok(string(output))
}

// HandleGeneratePlot calls the PaperBanana CLI to generate a statistical plot.
// Expected args:
//   - data_json (string): JSON string representing the data.
//   - intent (string): description of the desired plot.
//   - iterations (int, optional): number of refinement iterations (default 3).
func HandleGeneratePlot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dataJSON, _ :=getString(args, "data_json")
	if dataJSON == "" {
		return err("data_json is required")
}

	intent, _ :=getString(args, "intent")
	if intent == "" {
		return err("intent is required")
}

	iterations, _ :=getInt(args, "iterations")
	if iterations == 0 {
		iterations = 3
	}

	// Write data JSON to a temporary file
	tmpFile, tmpErr := ioutil.TempFile("", "paperbanana_data_*.json")
	if tmpErr != nil {
		return err(fmt.Sprintf("failed to create temp file for data: %v", tmpErr))
}

	defer os.Remove(tmpFile.Name())

	if _, writeErr := tmpFile.WriteString(dataJSON); writeErr != nil {
		return err(fmt.Sprintf("failed to write data json: %v", writeErr))
}

	if closeErr := tmpFile.Close(); closeErr != nil {
		return err(fmt.Sprintf("failed to close data temp file: %v", closeErr))
}

	cmdArgs := []string{
		"plot",
		"--data", tmpFile.Name(),
		"--intent", intent,
		"--iterations", strconv.Itoa(iterations),
	}
	cmd := exec.CommandContext(ctx, "paperbanana", cmdArgs...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("paperbanana plot failed: %v, output: %s", execErr, string(output)))
}

	return ok(string(output))
}