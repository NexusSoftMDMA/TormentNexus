package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleFilesystemTools(t *testing.T) {
	// Create temporary directory for testing
	tempDir, errMk := os.MkdirTemp("", "mcp-filesystem-test-*")
	if errMk != nil {
		t.Fatalf("Failed to create temp dir: %v", errMk)
	}
	defer os.RemoveAll(tempDir)

	ctx := context.Background()

	// 1. Test HandleCreateDirectory
	subDir := filepath.Join(tempDir, "subdir")
	resp, err := HandleCreateDirectory(ctx, map[string]interface{}{
		"path": subDir,
	})
	if err != nil {
		t.Fatalf("HandleCreateDirectory failed: %v", err)
	}
	if resp.IsError {
		t.Errorf("HandleCreateDirectory returned error: %s", resp.Content[0].Text)
	}

	// Verify directory exists
	if stat, errStat := os.Stat(subDir); errStat != nil || !stat.IsDir() {
		t.Errorf("Expected directory to exist at %s", subDir)
	}

	// 2. Test HandleWriteFile (from parity.go or standard mock write)
	testFile := filepath.Join(subDir, "test.txt")
	writeContent := "line1\nline2\nline3\nline4\nline5\nline6"
	resp, err = HandleWrite(ctx, map[string]interface{}{
		"file_path": testFile,
		"content":   writeContent,
	})
	if err != nil {
		t.Fatalf("HandleWrite failed: %v", err)
	}
	if resp.IsError {
		t.Errorf("HandleWrite returned error: %s", resp.Content[0].Text)
	}

	// 3. Test HandleReadTextFile (Complete)
	resp, err = HandleReadTextFile(ctx, map[string]interface{}{
		"path": testFile,
	})
	if err != nil {
		t.Fatalf("HandleReadTextFile failed: %v", err)
	}
	if resp.Content[0].Text != writeContent {
		t.Errorf("Expected complete read, got: %s", resp.Content[0].Text)
	}

	// 3b. Test HandleReadTextFile (Head)
	resp, err = HandleReadTextFile(ctx, map[string]interface{}{
		"path": testFile,
		"head": 2.0,
	})
	if err != nil {
		t.Fatalf("HandleReadTextFile (head) failed: %v", err)
	}
	if resp.Content[0].Text != "line1\nline2" {
		t.Errorf("Expected head output 'line1\\nline2', got: %q", resp.Content[0].Text)
	}

	// 3c. Test HandleReadTextFile (Tail)
	resp, err = HandleReadTextFile(ctx, map[string]interface{}{
		"path": testFile,
		"tail": 2.0,
	})
	if err != nil {
		t.Fatalf("HandleReadTextFile (tail) failed: %v", err)
	}
	if resp.Content[0].Text != "line5\nline6" {
		t.Errorf("Expected tail output 'line5\\nline6', got: %q", resp.Content[0].Text)
	}

	// 4. Test HandleListDirectory
	resp, err = HandleListDirectory(ctx, map[string]interface{}{
		"path": subDir,
	})
	if err != nil {
		t.Fatalf("HandleListDirectory failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "test.txt") {
		t.Errorf("Expected list_directory to contain test.txt, got: %s", resp.Content[0].Text)
	}

	// 5. Test HandleListDirectoryWithSizes
	resp, err = HandleListDirectoryWithSizes(ctx, map[string]interface{}{
		"path":   subDir,
		"sortBy": "size",
	})
	if err != nil {
		t.Fatalf("HandleListDirectoryWithSizes failed: %v", err)
	}
	var entries []FileEntry
	if errJson := json.Unmarshal([]byte(resp.Content[0].Text), &entries); errJson != nil {
		t.Fatalf("Failed to parse list_directory_with_sizes json: %v", errJson)
	}
	if len(entries) != 1 || entries[0].Name != "test.txt" || entries[0].Size != int64(len(writeContent)) {
		t.Errorf("Unexpected values in list_directory_with_sizes entries: %+v", entries)
	}

	// 6. Test HandleGetFileInfo
	resp, err = HandleGetFileInfo(ctx, map[string]interface{}{
		"path": testFile,
	})
	if err != nil {
		t.Fatalf("HandleGetFileInfo failed: %v", err)
	}
	var meta map[string]interface{}
	if errJson := json.Unmarshal([]byte(resp.Content[0].Text), &meta); errJson != nil {
		t.Fatalf("Failed to parse get_file_info json: %v", errJson)
	}
	if meta["name"] != "test.txt" || meta["is_directory"] != false {
		t.Errorf("Unexpected metadata values: %+v", meta)
	}

	// 7. Test HandleSearchFiles
	resp, err = HandleSearchFiles(ctx, map[string]interface{}{
		"path":    tempDir,
		"pattern": "*.txt",
	})
	if err != nil {
		t.Fatalf("HandleSearchFiles failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "test.txt") {
		t.Errorf("Expected search results to match test.txt, got: %s", resp.Content[0].Text)
	}

	// 8. Test HandleDirectoryTree
	resp, err = HandleDirectoryTree(ctx, map[string]interface{}{
		"path": tempDir,
	})
	if err != nil {
		t.Fatalf("HandleDirectoryTree failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "subdir/") || !strings.Contains(resp.Content[0].Text, "test.txt") {
		t.Errorf("Expected directory tree representation, got: %s", resp.Content[0].Text)
	}

	// 9. Test HandleMoveFile
	destFile := filepath.Join(tempDir, "moved.txt")
	resp, err = HandleMoveFile(ctx, map[string]interface{}{
		"source":      testFile,
		"destination": destFile,
	})
	if err != nil {
		t.Fatalf("HandleMoveFile failed: %v", err)
	}
	if stat, errStat := os.Stat(destFile); errStat != nil || stat.IsDir() {
		t.Errorf("Expected moved file to exist at %s", destFile)
	}
}
