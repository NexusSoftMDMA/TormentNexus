package mcpimpl

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestSerenaTools(t *testing.T) {
	// Create a temporary directory structure for tests
	tmpDir, err := os.MkdirTemp("", "serena-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Switch to the temp dir to run the tests
	oldCwd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}
	defer os.Chdir(oldCwd)

	// Write mock Go files
	mockGoCode := `package main

import "fmt"

// MessageStruct represents a simple struct
type MessageStruct struct {
	Text  string
	Count int
}

// GetText returns the text of the message
func (m *MessageStruct) GetText() string {
	return m.Text
}

// PrintMessage prints the message
func PrintMessage(msg string) {
	fmt.Println(msg)
}

const DefaultLimit = 100
var CurrentCount = 0
`

	relPath := "main.go"
	if err := os.WriteFile(relPath, []byte(mockGoCode), 0644); err != nil {
		t.Fatalf("failed to write mock file: %v", err)
	}

	ctx := context.Background()

	// 1. Test HandleGetSymbolsOverview
	t.Run("GetSymbolsOverview", func(t *testing.T) {
		args := map[string]interface{}{
			"relative_path": relPath,
			"depth":         1.0,
		}
		resp, err := HandleGetSymbolsOverview(ctx, args)
		if err != nil {
			t.Fatalf("HandleGetSymbolsOverview failed: %v", err)
		}
		if resp.IsError {
			t.Fatalf("HandleGetSymbolsOverview returned error: %s", resp.Content[0].Text)
		}

		var syms []map[string]interface{}
		if err := json.Unmarshal([]byte(resp.Content[0].Text), &syms); err != nil {
			t.Fatalf("failed to parse JSON response: %v", err)
		}

		// Expecting MessageStruct, GetText, and PrintMessage
		foundStruct := false
		foundFunc := false
		for _, s := range syms {
			name := s["name"].(string)
			if name == "MessageStruct" {
				foundStruct = true
				// Check fields are in children (since depth = 1)
				children := s["children"].([]interface{})
				if len(children) < 2 {
					t.Errorf("expected fields in children, got %v", children)
				}
			}
			if name == "PrintMessage" {
				foundFunc = true
			}
		}

		if !foundStruct {
			t.Errorf("expected MessageStruct in overview")
		}
		if !foundFunc {
			t.Errorf("expected PrintMessage in overview")
		}
	})

	// 2. Test HandleFindSymbol
	t.Run("FindSymbol", func(t *testing.T) {
		args := map[string]interface{}{
			"name_path_pattern": "MessageStruct/GetText",
			"depth":             0.0,
		}
		resp, err := HandleFindSymbol(ctx, args)
		if err != nil {
			t.Fatalf("HandleFindSymbol failed: %v", err)
		}
		if resp.IsError {
			t.Fatalf("HandleFindSymbol returned error: %s", resp.Content[0].Text)
		}

		if !strings.Contains(resp.Content[0].Text, "GetText") {
			t.Errorf("expected to find GetText method, got: %s", resp.Content[0].Text)
		}
	})

	// 3. Test HandleFindReferencingSymbols
	t.Run("FindReferencingSymbols", func(t *testing.T) {
		// Write a second mock file referencing main.go's functions
		mockGoCode2 := `package main

func UseMessage() {
	PrintMessage("hello")
}
`
		relPath2 := "use.go"
		if err := os.WriteFile(relPath2, []byte(mockGoCode2), 0644); err != nil {
			t.Fatalf("failed to write second mock file: %v", err)
		}

		args := map[string]interface{}{
			"name_path":     "PrintMessage",
			"relative_path": relPath,
		}
		resp, err := HandleFindReferencingSymbols(ctx, args)
		if err != nil {
			t.Fatalf("HandleFindReferencingSymbols failed: %v", err)
		}
		if resp.IsError {
			t.Fatalf("HandleFindReferencingSymbols returned error: %s", resp.Content[0].Text)
		}

		if !strings.Contains(resp.Content[0].Text, "use.go") {
			t.Errorf("expected reference to be found in use.go, got: %s", resp.Content[0].Text)
		}
	})

	// 4. Test HandleFindDeclaration
	t.Run("FindDeclaration", func(t *testing.T) {
		mockGoCode3 := `package main

func TargetFunc() {
	PrintMessage("val")
}
`
		relPath3 := "target.go"
		if err := os.WriteFile(relPath3, []byte(mockGoCode3), 0644); err != nil {
			t.Fatalf("failed to write third mock file: %v", err)
		}

		args := map[string]interface{}{
			"relative_path": relPath3,
			"regex":         `(PrintMessage)\(`,
		}
		resp, err := HandleFindDeclaration(ctx, args)
		if err != nil {
			t.Fatalf("HandleFindDeclaration failed: %v", err)
		}
		if resp.IsError {
			t.Fatalf("HandleFindDeclaration returned error: %s", resp.Content[0].Text)
		}

		if !strings.Contains(resp.Content[0].Text, "PrintMessage") {
			t.Errorf("expected PrintMessage declaration, got: %s", resp.Content[0].Text)
		}
	})

	// 5. Test HandleRenameSymbol
	t.Run("RenameSymbol", func(t *testing.T) {
		args := map[string]interface{}{
			"name_path":     "PrintMessage",
			"relative_path": relPath,
			"new_name":      "ShowMessage",
		}
		resp, err := HandleRenameSymbol(ctx, args)
		if err != nil {
			t.Fatalf("HandleRenameSymbol failed: %v", err)
		}
		if resp.IsError {
			t.Fatalf("HandleRenameSymbol returned error: %s", resp.Content[0].Text)
		}

		// Read file back and check it renamed
		data, _ := os.ReadFile(relPath)
		if !strings.Contains(string(data), "ShowMessage") {
			t.Errorf("expected main.go to contain ShowMessage, got: %s", string(data))
		}
	})
}
