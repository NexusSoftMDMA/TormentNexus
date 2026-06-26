package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func HandleParseThreadDump(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "file_path")
	if filePath == "" {
		return err("file_path is required")
}

	absPath, pathErr := filepath.Abs(filePath)
	if pathErr != nil {
		return err(fmt.Sprintf("failed to get absolute path: %v", pathErr))
}

	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		return err(fmt.Sprintf("file does not exist: %s", absPath))
}

	cmd := exec.Command("java", "-jar", "tda.jar", absPath)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("failed to execute TDA: %v, output: %s", execErr, string(output)))
}

	return ok(string(output))
}

func HandleAnalyzeDeadlocks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "file_path")
	if filePath == "" {
		return err("file_path is required")
}

	absPath, pathErr := filepath.Abs(filePath)
	if pathErr != nil {
		return err(fmt.Sprintf("failed to get absolute path: %v", pathErr))
}

	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		return err(fmt.Sprintf("file does not exist: %s", absPath))
}

	cmd := exec.Command("java", "-jar", "tda.jar", "--deadlocks", absPath)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("failed to analyze deadlocks: %v, output: %s", execErr, string(output)))
}

	return ok(string(output))
}

func HandleListThreads(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "file_path")
	if filePath == "" {
		return err("file_path is required")
}

	absPath, pathErr := filepath.Abs(filePath)
	if pathErr != nil {
		return err(fmt.Sprintf("failed to get absolute path: %v", pathErr))
}

	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		return err(fmt.Sprintf("file does not exist: %s", absPath))
}

	cmd := exec.Command("java", "-jar", "tda.jar", "--list-threads", absPath)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("failed to list threads: %v, output: %s", execErr, string(output)))
}

	return ok(string(output))
}

func HandleThreadStatistics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "file_path")
	if filePath == "" {
		return err("file_path is required")
}

	absPath, pathErr := filepath.Abs(filePath)
	if pathErr != nil {
		return err(fmt.Sprintf("failed to get absolute path: %v", pathErr))
}

	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		return err(fmt.Sprintf("file does not exist: %s", absPath))
}

	cmd := exec.Command("java", "-jar", "tda.jar", "--stats", absPath)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("failed to get thread statistics: %v, output: %s", execErr, string(output)))
}

	return ok(string(output))
}

func HandleFilterThreads(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "file_path")
	filter, _ :=getString(args, "filter")
	if filePath == "" {
		return err("file_path is required")
}

	if filter == "" {
		return err("filter is required")
}

	absPath, pathErr := filepath.Abs(filePath)
	if pathErr != nil {
		return err(fmt.Sprintf("failed to get absolute path: %v", pathErr))
}

	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		return err(fmt.Sprintf("file does not exist: %s", absPath))
}

	cmd := exec.Command("java", "-jar", "tda.jar", "--filter", filter, absPath)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("failed to filter threads: %v, output: %s", execErr, string(output)))
}

	return ok(string(output))
}