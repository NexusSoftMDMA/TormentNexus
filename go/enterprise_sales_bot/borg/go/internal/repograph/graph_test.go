package repograph

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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
		if edge.To == "import:fmt" || edge.To == "import:std/fmt" {
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

func TestRepoGraphTSResolution(t *testing.T) {
	tempDir := t.TempDir()

	// Create structure:
	// src/index.ts -> imports ./util
	// src/util.ts
	
	srcDir := filepath.Join(tempDir, "src")
	os.MkdirAll(srcDir, 0755)
	
	os.WriteFile(filepath.Join(srcDir, "util.ts"), []byte("export function util() {}"), 0644)
	os.WriteFile(filepath.Join(srcDir, "index.ts"), []byte("import { util } from './util'"), 0644)

	rgs := NewRepoGraphService(tempDir)
	graph, _ := rgs.Build(context.Background())

	foundResolved := false
	for _, edge := range graph.Edges {
		if edge.From == "file:src/index.ts" && edge.To == "file:src/util.ts" {
			foundResolved = true
		}
	}

	if !foundResolved {
		t.Error("TS relative import was not resolved to file:src/util.ts")
	}
}

func TestRepoGraphPythonResolution(t *testing.T) {
	tempDir := t.TempDir()

	// Create structure:
	// pkg/main.py -> from . import helper
	// pkg/helper.py
	
	pkgDir := filepath.Join(tempDir, "pkg")
	os.MkdirAll(pkgDir, 0755)
	
	os.WriteFile(filepath.Join(pkgDir, "helper.py"), []byte("def help(): pass"), 0644)
	os.WriteFile(filepath.Join(pkgDir, "main.py"), []byte("from . import helper"), 0644)

	rgs := NewRepoGraphService(tempDir)
	graph, _ := rgs.Build(context.Background())

	foundResolved := false
	for _, edge := range graph.Edges {
		if edge.From == "file:pkg/main.py" && edge.To == "file:pkg/helper.py" {
			foundResolved = true
		}
	}

	if !foundResolved {
		t.Logf("Edges found: %v", graph.Edges)
		t.Error("Python relative import was not resolved to file:pkg/helper.py")
	}
}

func TestRepoGraphPythonParentResolution(t *testing.T) {
	tempDir := t.TempDir()

	// Create structure:
	// pkg/sub/main.py -> from .. import base
	// pkg/base.py
	
	pkgDir := filepath.Join(tempDir, "pkg")
	subDir := filepath.Join(pkgDir, "sub")
	os.MkdirAll(subDir, 0755)
	
	os.WriteFile(filepath.Join(pkgDir, "base.py"), []byte("def base(): pass"), 0644)
	os.WriteFile(filepath.Join(subDir, "main.py"), []byte("from .. import base"), 0644)

	rgs := NewRepoGraphService(tempDir)
	graph, _ := rgs.Build(context.Background())

	foundResolved := false
	for _, edge := range graph.Edges {
		if edge.From == "file:pkg/sub/main.py" && edge.To == "file:pkg/base.py" {
			foundResolved = true
		}
	}

	if !foundResolved {
		t.Logf("Edges found: %v", graph.Edges)
		t.Error("Python parent relative import was not resolved to file:pkg/base.py")
	}
}

func TestRepoGraphPythonSiblingModResolution(t *testing.T) {
	tempDir := t.TempDir()

	// Create structure:
	// pkg/sub/main.py -> from ..other import tool
	// pkg/other.py
	
	pkgDir := filepath.Join(tempDir, "pkg")
	subDir := filepath.Join(pkgDir, "sub")
	os.MkdirAll(subDir, 0755)
	
	os.WriteFile(filepath.Join(pkgDir, "other.py"), []byte("def tool(): pass"), 0644)
	os.WriteFile(filepath.Join(subDir, "main.py"), []byte("from ..other import tool"), 0644)

	rgs := NewRepoGraphService(tempDir)
	graph, _ := rgs.Build(context.Background())

	foundResolved := false
	for _, edge := range graph.Edges {
		if edge.From == "file:pkg/sub/main.py" && edge.To == "file:pkg/other.py" {
			foundResolved = true
		}
	}

	if !foundResolved {
		t.Logf("Edges found: %v", graph.Edges)
		t.Error("Python sibling relative import was not resolved to file:pkg/other.py")
	}
}

func TestRepoGraphRustResolution(t *testing.T) {
	tempDir := t.TempDir()

	// Create structure:
	// src/lib.rs
	// src/models.rs
	// src/main.rs -> use crate::models; use super::something (invalid here but testable)
	
	srcDir := filepath.Join(tempDir, "src")
	os.MkdirAll(srcDir, 0755)
	
	os.WriteFile(filepath.Join(srcDir, "models.rs"), []byte("pub struct User {}"), 0644)
	os.WriteFile(filepath.Join(srcDir, "main.rs"), []byte("use crate::src::models;"), 0644)

	rgs := NewRepoGraphService(tempDir)
	graph, _ := rgs.Build(context.Background())

	foundResolved := false
	for _, edge := range graph.Edges {
		if edge.From == "file:src/main.rs" && edge.To == "file:src/models.rs" {
			foundResolved = true
		}
	}

	if !foundResolved {
		t.Logf("Edges found: %v", graph.Edges)
		t.Error("Rust 'use crate' import was not resolved to file:src/models.rs")
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

func TestRepoGraphCircularDependency(t *testing.T) {
	tempDir := t.TempDir()

	// Create structure:
	// a.go imports b.go
	// b.go imports a.go
	
	os.WriteFile(filepath.Join(tempDir, "a.go"), []byte("package a\nimport \"b\"\nfunc A() {}"), 0644)
	os.WriteFile(filepath.Join(tempDir, "b.go"), []byte("package b\nimport \"a\"\nfunc B() {}"), 0644)

	// Mock resolver to force circularity in test
	rgs := NewRepoGraphService(tempDir)
	graph, _ := rgs.Build(context.Background())
	
	// Manually add edges for test simplicity
	graph.Edges = append(graph.Edges, Edge{From: "file:a.go", To: "file:b.go", Type: "imports"})
	graph.Edges = append(graph.Edges, Edge{From: "file:b.go", To: "file:a.go", Type: "imports"})

	cycles := rgs.GetCircularDependencies()
	if len(cycles) == 0 {
		t.Error("Failed to detect circular dependency")
	}
}

func TestRepoGraphImpactAnalysis(t *testing.T) {
	tempDir := t.TempDir()

	// Create chain: a -> b -> c
	os.WriteFile(filepath.Join(tempDir, "c.go"), []byte("package c"), 0644)
	os.WriteFile(filepath.Join(tempDir, "b.go"), []byte("package b\nimport \"c\""), 0644)
	os.WriteFile(filepath.Join(tempDir, "a.go"), []byte("package a\nimport \"b\""), 0644)

	rgs := NewRepoGraphService(tempDir)
	graph, _ := rgs.Build(context.Background())
	
	// Mock resolution
	graph.Edges = append(graph.Edges, Edge{From: "file:b.go", To: "file:c.go", Type: "imports"})
	graph.Edges = append(graph.Edges, Edge{From: "file:a.go", To: "file:b.go", Type: "imports"})

	impacted := rgs.GetImpactAnalysis("c.go")
	
	foundA := false
	foundB := false
	for _, p := range impacted {
		if p == "a.go" { foundA = true }
		if p == "b.go" { foundB = true }
	}

	if !foundA || !foundB {
		t.Errorf("Impact analysis failed, expected [a.go b.go], got %v", impacted)
	}
}

func TestRepoGraphFindDefinitions(t *testing.T) {
	tempDir := t.TempDir()

	os.WriteFile(filepath.Join(tempDir, "a.go"), []byte("package a\nfunc MyFunc() {}"), 0644)

	rgs := NewRepoGraphService(tempDir)
	_, _ = rgs.Build(context.Background())

	defs := rgs.FindDefinitions("MyFunc")
	if len(defs) == 0 || defs[0].Name != "MyFunc" {
		t.Error("Failed to find definition for MyFunc")
	}
}

func TestRepoGraphFindReferences(t *testing.T) {
	tempDir := t.TempDir()

	os.WriteFile(filepath.Join(tempDir, "a.go"), []byte("package a\nfunc MyFunc() {}"), 0644)

	rgs := NewRepoGraphService(tempDir)
	graph, _ := rgs.Build(context.Background())
	
	// Manually add a call reference for test
	// We need to ensure the target node exists in the graph nodes map
	// The node MUST have the correct type and name to be found as a target
	graph.Nodes["a.go#MyFunc"] = &Node{ID: "a.go#MyFunc", Name: "MyFunc", Type: NodeFunction}
	graph.Nodes["file:b.go"] = &Node{ID: "file:b.go", Name: "b.go", Type: NodeFile}
	graph.Edges = append(graph.Edges, Edge{From: "file:b.go", To: "a.go#MyFunc", Type: "calls"})

	refs := rgs.FindReferences("MyFunc")
	if len(refs) == 0 {
		t.Error("Failed to find references for MyFunc")
	}
}

func TestRepoGraphDynamicGoModuleDetection(t *testing.T) {
	tempDir := t.TempDir()

	// Create a go.mod with a custom module path
	goMod := "module custom.io/my-project\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a Go file importing from the custom module
	goContent := `package main

import (
	"fmt"
	"custom.io/my-project/internal/config"
	"github.com/external/dep"
)

func main() {
	fmt.Println("hello")
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}

	rgs := NewRepoGraphService(tempDir)
	graph, err := rgs.Build(context.Background())
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify module detection
	modulePath := rgs.detectGoModule()
	if modulePath != "custom.io/my-project" {
		t.Errorf("expected module path 'custom.io/my-project', got %q", modulePath)
	}

	// Check that stdlib imports are categorized correctly
	hasStdlib := false
	hasExternal := false
	for _, edge := range graph.Edges {
		if edge.To == "import:std/fmt" {
			hasStdlib = true
		}
		if strings.HasPrefix(edge.To, "import:github/") {
			hasExternal = true
		}
	}
	if !hasStdlib {
		t.Error("expected stdlib import edge for fmt")
	}
	if !hasExternal {
		t.Error("expected external import edge for github.com/external/dep")
	}
}

func TestRepoGraphExternalImportCategorization(t *testing.T) {
	rgs := NewRepoGraphService(t.TempDir())

	tests := []struct {
		importPath string
		expected   string
	}{
		{"github.com/some/repo", "import:github/some/repo"},
		{"gitlab.com/group/project", "import:gitlab/group/project"},
		{"golang.org/x/text", "import:golang-x/text"},
		{"cloud.google.com/go/storage", "import:gcp/go/storage"},
		{"k8s.io/client-go", "import:k8s/client-go"},
		{"gopkg.in/yaml.v3", "import:gopkg/yaml.v3"},
		{"go.uber.org/zap", "import:uber/zap"},
		{"some-unknown.host/pkg", "import:external/some-unknown.host/pkg"},
	}

	for _, tt := range tests {
		result := rgs.categorizeExternalImport(tt.importPath)
		if result != tt.expected {
			t.Errorf("categorizeExternalImport(%q) = %q, want %q", tt.importPath, result, tt.expected)
		}
	}
}
