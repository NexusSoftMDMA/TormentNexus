package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleGitIngest_Local(t *testing.T) {
	// Create a temp directory structure for testing
	tempDir, err := os.MkdirTemp("", "gitingest-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirs and files
	subDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	file1 := filepath.Join(tempDir, "README.md")
	if err := os.WriteFile(file1, []byte("# Test Repo\nThis is a test readme."), 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	file2 := filepath.Join(subDir, "main.go")
	if err := os.WriteFile(file2, []byte("package main\n\nfunc main() {}"), 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	file3 := filepath.Join(subDir, "binary.bin")
	if err := os.WriteFile(file3, []byte{0, 1, 2, 3, 4}, 0644); err != nil {
		t.Fatalf("Failed to write file3: %v", err)
	}

	// Run HandleGitIngest on local temp dir
	args := map[string]interface{}{
		"source":        tempDir,
		"max_file_size": 1024.0,
	}

	resp, errIngest := HandleGitIngest(context.Background(), args)
	if errIngest != nil {
		t.Fatalf("HandleGitIngest returned error: %v", errIngest)
	}

	if resp.IsError {
		t.Fatalf("HandleGitIngest response contains error: %s", resp.Content[0].Text)
	}

	outputText := resp.Content[0].Text

	// Verify content
	if !strings.Contains(outputText, "Total Files: 3") {
		t.Errorf("Expected total files to be 3, output: %s", outputText)
	}

	if !strings.Contains(outputText, "README.md") || !strings.Contains(outputText, "main.go") {
		t.Errorf("Expected README.md and main.go in output, got: %s", outputText)
	}

	if !strings.Contains(outputText, "[Binary File]") {
		t.Errorf("Expected binary file detection in output, got: %s", outputText)
	}
}

func TestHandleGitIngest_Filters(t *testing.T) {
	// Create a temp directory structure for testing
	tempDir, err := os.MkdirTemp("", "gitingest-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	file1 := filepath.Join(tempDir, "README.md")
	os.WriteFile(file1, []byte("# Test Repo"), 0644)

	file2 := filepath.Join(tempDir, "main.go")
	os.WriteFile(file2, []byte("package main"), 0644)

	// Filter only .go files
	args := map[string]interface{}{
		"source":           tempDir,
		"include_patterns": "main.go",
	}

	resp, errIngest := HandleGitIngest(context.Background(), args)
	if errIngest != nil {
		t.Fatalf("HandleGitIngest returned error: %v", errIngest)
	}

	outputText := resp.Content[0].Text
	if strings.Contains(outputText, "File: README.md") {
		t.Errorf("Expected README.md to be filtered out of file contents, got: %s", outputText)
	}
	if !strings.Contains(outputText, "File: main.go") {
		t.Errorf("Expected main.go to be included in file contents, got: %s", outputText)
	}
}

