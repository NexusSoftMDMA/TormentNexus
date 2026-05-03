package repograph

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRepoGraphBuild(t *testing.T) {
	tempDir := t.TempDir()

	// Create a dummy Go file
	goFile := filepath.Join(tempDir, "main.go")
	goContent := `package main
import "fmt"
func Main() {
	fmt.Println("Hello")
}
type Config struct {
	ID string
}
`
	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a dummy TS file
	tsFile := filepath.Join(tempDir, "index.ts")
	tsContent := `import { some } from "./other";
export function hello() {
	return "world";
}
export interface User {
	id: string;
}
`
	if err := os.WriteFile(tsFile, []byte(tsContent), 0644); err != nil {
		t.Fatal(err)
	}

	rgs := NewRepoGraphService(tempDir)
	graph, err := rgs.Build(context.Background())
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if graph.Stats.TotalFiles != 2 {
		t.Errorf("Expected 2 files, got %d", graph.Stats.TotalFiles)
	}

	// Verify Go symbols
	foundMain := false
	foundConfig := false
	for _, node := range graph.Nodes {
		if node.Name == "Main" && node.Type == NodeFunction && node.Language == "go" {
			foundMain = true
		}
		if node.Name == "Config" && node.Type == NodeTypeName && node.Language == "go" {
			foundConfig = true
		}
	}

	if !foundMain {
		t.Error("Did not find Go function 'Main'")
	}
	if !foundConfig {
		t.Error("Did not find Go type 'Config'")
	}

	// Verify TS symbols
	foundHello := false
	foundUser := false
	for _, node := range graph.Nodes {
		if node.Name == "hello" && node.Type == NodeFunction && node.Language == "typescript" {
			foundHello = true
		}
		if node.Name == "User" && node.Type == NodeInterface && node.Language == "typescript" {
			foundUser = true
		}
	}

	if !foundHello {
		t.Error("Did not find TS function 'hello'")
	}
	if !foundUser {
		t.Error("Did not find TS interface 'User'")
	}

	// Verify imports
	foundGoImport := false
	foundTSImport := false
	for _, edge := range graph.Edges {
		if edge.To == "import:fmt" {
			foundGoImport = true
		}
		if edge.To == "import:./other" {
			foundTSImport = true
		}
	}

	if !foundGoImport {
		t.Error("Did not find Go import 'fmt'")
	}
	if !foundTSImport {
		t.Error("Did not find TS import './other'")
	}
}

func TestRepoGraphSearch(t *testing.T) {
	tempDir := t.TempDir()
	os.WriteFile(filepath.Join(tempDir, "a.go"), []byte("package a\nfunc SearchMe() {}"), 0644)
	
	rgs := NewRepoGraphService(tempDir)
	_, _ = rgs.Build(context.Background())

	results := rgs.SearchSymbols("Search", 10)
	if len(results) != 1 || results[0].Name != "SearchMe" {
		t.Errorf("Search failed, got %v", results)
	}
}
