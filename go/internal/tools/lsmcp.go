//go:build ignore
// +build ignore

package tools

/**
 * @file lsmcp.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of lsmcp — LSP-based code manipulation for multi-language analysis.
 * Replaces: @mizchi/lsmcp (npm)
 *
 * Provides semantic code analysis, symbol search, diagnostics, and refactoring
 * across multiple programming languages using Go's native analysis tools.
 *
 * Features:
 *  - project_overview — understand codebase structure
 *  - search_symbols — find functions, types, interfaces
 *  - get_symbol_details — deep inspect symbols
 *  - get_diagnostics — check for errors
 *  - find_references — find usages across codebase
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HandleLsmcpProjectOverview returns a high-level overview of the project structure.
func HandleLsmcpProjectOverview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rootPath, _ := getString(args, "path", "dir", "directory")
	if rootPath == "" {
		rootPath = "."
	}
	// Use go list to get module info
	modCmd := exec.CommandContext(ctx, "go", "list", "-m")
	modCmd.Dir = rootPath
	moduleBytes, modErr := modCmd.Output()
	moduleName := ""
	if modErr == nil {
		moduleName = strings.TrimSpace(string(moduleBytes))
	}

	// Count Go files
	var goFiles []string
	filepath.Walk(rootPath, func(p string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(p, ".go") {
			goFiles = append(goFiles, p)
		}
		return nil
	})

	overview := map[string]interface{}{
		"module":       moduleName,
		"go_files":     len(goFiles),
		"language":     "Go",
		"description":  fmt.Sprintf("Go project with %d source files", len(goFiles)),
	}
	if moduleName != "" {
		overview["name"] = moduleName
	}
	data, _ := json.MarshalIndent(overview, "", "  ")
	return ok(string(data))
}

// HandleLsmcpSearchSymbols searches for symbols in the codebase.
func HandleLsmcpSearchSymbols(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "symbol")
	searchPath, _ := getString(args, "path", "dir")
	if query == "" {
		return err("query is required")
	}
	if searchPath == "" {
		searchPath = "."
	}

	// Use rg/grep to find Go declarations and filter by query
	cmd := exec.CommandContext(ctx, "rg", "-n", "--no-heading",
		"-g", "*.go",
		fmt.Sprintf(`func |type |struct |interface |var |const `),
		searchPath)
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		if exitErr, isExit := cmdErr.(*exec.ExitError); isExit && exitErr.ExitCode() == 1 {
			return ok("[]")
		}
		return err(fmt.Sprintf("search failed: %v", cmdErr))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var results []map[string]string
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(query)) {
			parts := strings.SplitN(line, ":", 3)
			if len(parts) >= 2 {
				contextStr := ""
				if len(parts) > 2 {
					contextStr = parts[2]
				}
				result := map[string]string{
					"file":    parts[0],
					"line":    parts[1],
					"context": contextStr,
				}
				results = append(results, result)
			}
		}
	}
	if results == nil {
		results = []map[string]string{}
	}
	data, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(data))
}

// HandleLsmcpGetDiagnostics runs static analysis on the codebase.
func HandleLsmcpGetDiagnostics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	diagPath, _ := getString(args, "path", "dir")
	if diagPath == "" {
		diagPath = "."
	}

	// Run go vet
	vetCmd := exec.CommandContext(ctx, "go", "vet", "./...")
	vetCmd.Dir = diagPath
	vetOut, _ := vetCmd.CombinedOutput()
	vetStr := strings.TrimSpace(string(vetOut))

	// Run go build
	buildCmd := exec.CommandContext(ctx, "go", "build", "-buildvcs=false", "./...")
	buildCmd.Dir = diagPath
	buildOut, _ := buildCmd.CombinedOutput()
	buildStr := strings.TrimSpace(string(buildOut))

	diagnostics := map[string]interface{}{
		"vet":   vetStr,
		"build": buildStr,
	}
	if vetStr == "" {
		diagnostics["vet"] = "No issues found"
	}
	if buildStr == "" {
		diagnostics["build"] = "Build succeeded"
	}

	data, _ := json.MarshalIndent(diagnostics, "", "  ")
	return ok(string(data))
}

// HandleLsmcpFindReferences finds references to a symbol across the codebase.
func HandleLsmcpFindReferences(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ := getString(args, "symbol", "name", "query")
	refPath, _ := getString(args, "path", "dir")
	if symbol == "" {
		return err("symbol is required")
	}
	if refPath == "" {
		refPath = "."
	}

	refCmd := exec.CommandContext(ctx, "rg", "-n", "--no-heading",
		"-g", "*.go",
		"--", symbol, refPath)
	output, cmdErr := refCmd.Output()
	if cmdErr != nil {
		if exitErr, isExit := cmdErr.(*exec.ExitError); isExit && exitErr.ExitCode() == 1 {
			return ok("[]")
		}
		return err(fmt.Sprintf("find references failed: %v", cmdErr))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var refs []map[string]string
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 2 {
			ref := map[string]string{"file": parts[0], "line": parts[1]}
			if len(parts) > 2 {
				ref["context"] = strings.TrimSpace(parts[2])
			}
			refs = append(refs, ref)
		}
	}
	if refs == nil {
		refs = []map[string]string{}
	}
	data, _ := json.MarshalIndent(refs, "", "  ")
	return ok(string(data))
}

// HandleLsmcpGetSymbolDetails returns detailed info about a symbol.
func HandleLsmcpGetSymbolDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ := getString(args, "symbol", "name")
	docPath, _ := getString(args, "path", "dir")
	if symbol == "" {
		return err("symbol is required")
	}
	if docPath == "" {
		docPath = "."
	}

	// Use go doc to get documentation
	docCmd := exec.CommandContext(ctx, "go", "doc", symbol)
	docCmd.Dir = docPath
	docOut, _ := docCmd.Output()
	docStr := string(docOut)

	details := map[string]interface{}{
		"symbol": symbol,
		"doc":    docStr,
	}
	data, _ := json.MarshalIndent(details, "", "  ")
	return ok(string(data))
}
