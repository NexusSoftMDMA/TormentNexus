package repograph

/**
 * @file graph.go
 * @module go/internal/repograph
 *
 * WHAT: Repository graph service — builds a dependency/import graph of source code,
 *       enabling semantic code navigation, impact analysis, and code search.
 *
 * WHY: Code understanding is fundamental to AI-assisted development. Knowing which
 *      files import which, what functions are defined where, and what the dependency
 *      tree looks like enables intelligent code suggestions and refactoring.
 *
 * ADDED: v1.0.0-alpha.32
 */

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// NodeType represents the type of a graph node.
type NodeType string

const (
	NodeFile       NodeType = "file"
	NodeFunction   NodeType = "function"
	NodeTypeName   NodeType = "type"
	NodeInterface  NodeType = "interface"
	NodeImport     NodeType = "import"
	NodePackage    NodeType = "package"
)

// Node represents a node in the repository graph.
type Node struct {
	ID          string   `json:"id"`
	Type        NodeType `json:"type"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Package     string   `json:"package,omitempty"`
	Language    string   `json:"language,omitempty"`
	LineStart   int      `json:"lineStart,omitempty"`
	LineEnd     int      `json:"lineEnd,omitempty"`
	IsExported  bool     `json:"isExported"`
	DocComment  string   `json:"docComment,omitempty"`
}

// Edge represents a dependency between two nodes.
type Edge struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Type    string `json:"type"` // "imports", "calls", "implements", "extends", "references"
	Weight  int    `json:"weight"`
}

// Graph is the complete repository dependency graph.
type Graph struct {
	Nodes    map[string]*Node `json:"nodes"`
	Edges    []Edge           `json:"edges"`
	RootPath string           `json:"rootPath"`
	BuiltAt  time.Time        `json:"builtAt"`
	Stats    GraphStats       `json:"stats"`
}

// GraphStats contains summary statistics about the graph.
type GraphStats struct {
	TotalFiles     int `json:"totalFiles"`
	TotalFunctions int `json:"totalFunctions"`
	TotalTypes     int `json:"totalTypes"`
	TotalImports   int `json:"totalImports"`
	TotalEdges     int `json:"totalEdges"`
}

// RepoGraphService builds and queries repository graphs.
type RepoGraphService struct {
	goModule   string
	goModuleMu sync.Once
	root  string
	mu    sync.RWMutex
	graph *Graph
}

// NewRepoGraphService creates a new graph service for the given root.
func NewRepoGraphService(root string) *RepoGraphService {
	return &RepoGraphService{
		root: root,
	}
}

// Build scans the repository and builds the dependency graph.
func (rgs *RepoGraphService) Build(ctx context.Context) (*Graph, error) {
	graph := &Graph{
		Nodes:    make(map[string]*Node),
		RootPath: rgs.root,
		BuiltAt:  time.Now().UTC(),
	}

	// 1. First pass: index all files to build the node map
	err := filepath.WalkDir(rgs.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			name := strings.ToLower(d.Name())
			switch name {
			case "node_modules", ".git", "dist", "build", "coverage", "vendor",
				"__pycache__", ".next", ".cache", "target", "bin":
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".rs":
			relPath, _ := filepath.Rel(rgs.root, path)
			relPath = filepath.ToSlash(relPath)
			fileID := "file:" + relPath
			graph.Nodes[fileID] = &Node{
				ID:       fileID,
				Type:     NodeFile,
				Name:     filepath.Base(path),
				Path:     relPath,
				Language: languageFromExt(ext),
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	rgs.mu.Lock()
	rgs.graph = graph
	rgs.mu.Unlock()

	// 2. Second pass: parse content and resolve imports
	err = filepath.WalkDir(rgs.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".rs":
			relPath, _ := filepath.Rel(rgs.root, path)
			relPath = filepath.ToSlash(relPath)
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			if len(data) > 500*1024 {
				return nil
			}

			switch ext {
			case ".go":
				rgs.indexGoFile(graph, relPath, string(data))
			case ".ts", ".tsx", ".js", ".jsx":
				rgs.indexTSFile(graph, relPath, string(data))
			case ".py":
				rgs.indexPythonFile(graph, relPath, string(data))
			case ".rs":
				rgs.indexRustFile(graph, relPath, string(data))
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Calculate stats
	graph.Stats = GraphStats{
		TotalFiles:     rgs.countByType(graph, NodeFile),
		TotalFunctions: rgs.countByType(graph, NodeFunction),
		TotalTypes:     rgs.countByType(graph, NodeTypeName),
		TotalImports:   rgs.countByType(graph, NodeImport),
		TotalEdges:     len(graph.Edges),
	}

	return graph, nil
}

func (rgs *RepoGraphService) GetNodeByID(id string) *Node {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()
	if rgs.graph == nil {
		return nil
	}
	return rgs.graph.Nodes[id]
}

func (rgs *RepoGraphService) GetStats() GraphStats {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()
	if rgs.graph == nil {
		return GraphStats{}
	}
	return rgs.graph.Stats
}

func (rgs *RepoGraphService) FindDefinitions(symbolName string) []*Node {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()

	if rgs.graph == nil {
		return nil
	}

	var defs []*Node
	for _, node := range rgs.graph.Nodes {
		if node.Type != NodeFile && node.Type != NodeImport && node.Name == symbolName {
			defs = append(defs, node)
		}
	}
	return defs
}

func (rgs *RepoGraphService) GetImpactAnalysis(filePath string) []string {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()

	if rgs.graph == nil {
		return nil
	}

	fileID := "file:" + filePath
	impacted := make(map[string]bool)
	var queue []string

	// Start with files that directly import the target file
	for _, edge := range rgs.graph.Edges {
		if edge.To == fileID && edge.Type == "imports" {
			if _, exists := impacted[edge.From]; !exists {
				impacted[edge.From] = true
				queue = append(queue, edge.From)
			}
		}
	}

	// Breadth-First Search to find all transitive dependents
	head := 0
	for head < len(queue) {
		current := queue[head]
		head++

		for _, edge := range rgs.graph.Edges {
			if edge.To == current && edge.Type == "imports" {
				if _, exists := impacted[edge.From]; !exists {
					impacted[edge.From] = true
					queue = append(queue, edge.From)
				}
			}
		}
	}

	var results []string
	for id := range impacted {
		if node, ok := rgs.graph.Nodes[id]; ok && node.Type == NodeFile {
			results = append(results, node.Path)
		}
	}
	sort.Strings(results)
	return results
}

func (rgs *RepoGraphService) GetCircularDependencies() [][]string {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()

	if rgs.graph == nil {
		return nil
	}

	var cycles [][]string
	visited := make(map[string]bool)
	stack := make(map[string]bool)
	var path []string

	var dfs func(u string)
	dfs = func(u string) {
		visited[u] = true
		stack[u] = true
		path = append(path, u)

		for _, edge := range rgs.graph.Edges {
			if edge.From == u && edge.Type == "imports" {
				v := edge.To
				if !visited[v] {
					dfs(v)
				} else if stack[v] {
					// Cycle detected
					var cycle []string
					found := false
					for _, node := range path {
						if node == v {
							found = true
						}
						if found {
							cycle = append(cycle, node)
						}
					}
					cycle = append(cycle, v)
					cycles = append(cycles, cycle)
				}
			}
		}

		stack[u] = false
		path = path[:len(path)-1]
	}

	for id, node := range rgs.graph.Nodes {
		if node.Type == NodeFile && !visited[id] {
			dfs(id)
		}
	}

	return cycles
}

// GetGraph returns the current graph, building it if necessary.
func (rgs *RepoGraphService) GetGraph() *Graph {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()
	return rgs.graph
}

// FindReferences finds all nodes that reference the given symbol.
func (rgs *RepoGraphService) FindReferences(symbolName string) []*Node {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()

	if rgs.graph == nil {
		return nil
	}

	var refs []*Node
	// 1. Find the target node(s) with this name (the definitions)
	var targets []string
	for id, node := range rgs.graph.Nodes {
		if node.Type != NodeFile && node.Type != NodeImport && node.Name == symbolName {
			targets = append(targets, id)
		}
	}

	if len(targets) == 0 {
		return nil
	}

	// 2. Find all edges that point to these targets
	seen := make(map[string]bool)
	for _, targetID := range targets {
		for _, edge := range rgs.graph.Edges {
			if edge.To == targetID {
				if fromNode, ok := rgs.graph.Nodes[edge.From]; ok {
					if !seen[edge.From] {
						refs = append(refs, fromNode)
						seen[edge.From] = true
					}
				}
			}
		}
	}
	return refs
}

// FindDependents returns all files that depend on the given file.
func (rgs *RepoGraphService) FindDependents(filePath string) []string {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()

	if rgs.graph == nil {
		return nil
	}

	fileID := "file:" + filePath
	var dependents []string
	seen := make(map[string]bool)

	var walk func(id string)
	walk = func(id string) {
		for _, edge := range rgs.graph.Edges {
			if edge.To == id && !seen[edge.From] {
				seen[edge.From] = true
				if node := rgs.graph.Nodes[edge.From]; node != nil && node.Type == NodeFile {
					dependents = append(dependents, node.Path)
				}
				walk(edge.From)
			}
		}
	}

	walk(fileID)
	return dependents
}

// SearchSymbols finds symbols matching a query.
func (rgs *RepoGraphService) SearchSymbols(query string, limit int) []*Node {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()

	if rgs.graph == nil {
		return nil
	}

	q := strings.ToLower(query)
	var results []*Node

	for _, node := range rgs.graph.Nodes {
		if node.Type == NodeFile || node.Type == NodeImport {
			continue
		}
		if strings.Contains(strings.ToLower(node.Name), q) {
			results = append(results, node)
		}
		if len(results) >= limit {
			break
		}
	}

	return results
}

var (
	goFuncRe    = regexp.MustCompile(`^func\s+(?:\([^)]+\)\s+)?(\w+)`)
	goTypeRe    = regexp.MustCompile(`^type\s+(\w+)\s+(struct|interface)`)
	goImportRe  = regexp.MustCompile(`^\s*"([^"]+)"`)
	goPackageRe = regexp.MustCompile(`^package\s+(\w+)`)

	tsFuncRe      = regexp.MustCompile(`(?:export\s+)?(?:async\s+)?function\s+(\w+)`)
	tsClassRe     = regexp.MustCompile(`(?:export\s+)?(?:abstract\s+)?class\s+(\w+)`)
	tsInterfaceRe = regexp.MustCompile(`(?:export\s+)?interface\s+(\w+)`)
	tsImportRe    = regexp.MustCompile(`import.*from\s+['"]([^'"]+)['"]`)

	pyFuncRe   = regexp.MustCompile(`^def\s+(\w+)`)
	pyClassRe  = regexp.MustCompile(`^class\s+(\w+)`)
	pyImportRe = regexp.MustCompile(`^import\s+([\w\.]+)|^from\s+([\w\.]+)\s+import\s+([\w\.]+)|^from\s+(\.)\s+import\s+([\w\.]+)`)

	rsFuncRe   = regexp.MustCompile(`(?:pub\s+)?fn\s+(\w+)`)
	rsStructRe = regexp.MustCompile(`(?:pub\s+)?struct\s+(\w+)`)
	rsModRe    = regexp.MustCompile(`(?:pub\s+)?mod\s+(\w+);`)
	rsUseRe    = regexp.MustCompile(`^use\s+([\w\:]+)(?::{([\w\s,]+)})?;`)

	tsAbsoluteImportRe = regexp.MustCompile(`^@/|^\w`)
	// RsAbsoluteImportRe helps identify external crates vs internal modules
	rsAbsoluteImportRe = regexp.MustCompile(`^(?:std|core|alloc|serde|tokio|anyhow|axum|reqwest|sqlx)::`)
)

func (rgs *RepoGraphService) resolveTSImport(currentFile, importPath string) string {
	if tsAbsoluteImportRe.MatchString(importPath) {
		return "import:" + importPath
	}

	dir := filepath.Dir(currentFile)
	target := filepath.ToSlash(filepath.Clean(filepath.Join(dir, importPath)))

	// Try extensions
	extensions := []string{".ts", ".tsx", ".js", ".jsx", "/index.ts", "/index.tsx"}
	for _, ext := range extensions {
		fullPath := target + ext
		if rgs.graph != nil {
			if _, ok := rgs.graph.Nodes["file:"+fullPath]; ok {
				return "file:" + fullPath
			}
		}
	}

	return "import:" + importPath
}

func (rgs *RepoGraphService) resolvePythonRelativeImport(relPath, dots, modName string) string {
	dir := filepath.Dir(relPath)
	if dir == "." {
		dir = ""
	}

	// Calculate how many levels to go up
	// dots is something like ".", "..", "..."
	levels := len(dots)
	current := dir
	for i := 1; i < levels; i++ {
		if current == "" || current == "." {
			break
		}
		current = filepath.Dir(current)
		if current == "." {
			current = ""
		}
	}

	targetPath := strings.ReplaceAll(modName, ".", "/")
	fullTarget := targetPath
	if current != "" {
		fullTarget = filepath.ToSlash(filepath.Clean(filepath.Join(current, targetPath)))
	}

	extensions := []string{".py", "/__init__.py"}
	for _, ext := range extensions {
		fullPath := fullTarget + ext
		if rgs.graph != nil {
			if _, ok := rgs.graph.Nodes["file:"+fullPath]; ok {
				return "file:" + fullPath
			}
		}
	}

	// Special case: if we are importing a module name that is a file in that target dir
	if rgs.graph != nil && current != "" {
		checkPath := filepath.ToSlash(filepath.Join(current, modName+".py"))
		if _, ok := rgs.graph.Nodes["file:"+checkPath]; ok {
			return "file:" + checkPath
		}
	}

	return "import:" + dots + modName
}

func (rgs *RepoGraphService) indexGoFile(graph *Graph, relPath, content string) {
	fileID := "file:" + relPath
	pkgName := ""

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	inImportBlock := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Package
		if matches := goPackageRe.FindStringSubmatch(line); len(matches) > 1 {
			pkgName = matches[1]
			if file := graph.Nodes[fileID]; file != nil {
				file.Package = pkgName
			}
		}

		// Imports
		if strings.HasPrefix(strings.TrimSpace(line), "import (") {
			inImportBlock = true
			continue
		}
		if inImportBlock {
			if strings.TrimSpace(line) == ")" {
				inImportBlock = false
				continue
			}
			if matches := goImportRe.FindStringSubmatch(line); len(matches) > 1 {
				target := rgs.resolveGoImport(matches[1])
				graph.Edges = append(graph.Edges, Edge{
					From: fileID, To: target, Type: "imports",
				})
			}
			continue
		}
		if matches := regexp.MustCompile(`import\s+"([^"]+)"`).FindStringSubmatch(line); len(matches) > 1 {
			target := rgs.resolveGoImport(matches[1])
			graph.Edges = append(graph.Edges, Edge{
				From: fileID, To: target, Type: "imports",
			})
		}

		// Functions
		if matches := goFuncRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			exported := name[0] >= 'A' && name[0] <= 'Z'
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeFunction,
				Name:       name,
				Path:       relPath,
				Package:    pkgName,
				LineStart:  lineNum,
				IsExported: exported,
				Language:   "go",
			}
		}

		// Types
		if matches := goTypeRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			nodeType := NodeTypeName
			if matches[2] == "interface" {
				nodeType = NodeInterface
			}
			exported := name[0] >= 'A' && name[0] <= 'Z'
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       nodeType,
				Name:       name,
				Path:       relPath,
				Package:    pkgName,
				LineStart:  lineNum,
				IsExported: exported,
				Language:   "go",
			}
		}
	}
}

func (rgs *RepoGraphService) indexTSFile(graph *Graph, relPath, content string) {
	fileID := "file:" + relPath

	// Imports
	for _, match := range tsImportRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			target := rgs.resolveTSImport(relPath, match[1])
			graph.Edges = append(graph.Edges, Edge{
				From: fileID, To: target, Type: "imports",
			})
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matches := tsFuncRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			exported := strings.Contains(line, "export")
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeFunction,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: exported,
				Language:   "typescript",
			}
		}

		if matches := tsClassRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeTypeName,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: strings.Contains(line, "export"),
				Language:   "typescript",
			}
		}

		if matches := tsInterfaceRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeInterface,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: strings.Contains(line, "export"),
				Language:   "typescript",
			}
		}
	}
}

func (rgs *RepoGraphService) indexPythonFile(graph *Graph, relPath, content string) {
	fileID := "file:" + relPath

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines or comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Generalized relative import regex: from ... import X
		relMultiDotRe := regexp.MustCompile(`^from\s+(\.+)([\w\.]*)\s+import\s+([\w\.,\s\(\)]+)`)
		if matches := relMultiDotRe.FindStringSubmatch(line); len(matches) > 3 {
			dots := matches[1]
			modName := matches[2]
			if modName != "" && strings.HasPrefix(modName, ".") {
				// handle from ..mod import x
				modName = strings.TrimPrefix(modName, ".")
			}
			symbolList := matches[3]
			
			// Clean symbol list (handle 'import (a, b)' or 'import a, b')
			symbolList = strings.Trim(symbolList, "()")
			symbols := strings.Split(symbolList, ",")
			
			for _, s := range symbols {
				symbol := strings.TrimSpace(s)
				if symbol == "" {
					continue
				}
				// Handle 'as' alias
				if strings.Contains(symbol, " as ") {
					symbol = strings.Fields(symbol)[0]
				}

				effectiveMod := modName
				if effectiveMod == "" {
					effectiveMod = symbol
				}

				target := rgs.resolvePythonRelativeImport(relPath, dots, effectiveMod)
				graph.Edges = append(graph.Edges, Edge{
					From: fileID, To: target, Type: "imports",
				})
			}
			continue
		}

		if matches := pyImportRe.FindStringSubmatch(line); len(matches) > 0 {
			mod := ""
			if matches[1] != "" { // import x
				mod = matches[1]
			} else if matches[2] != "" { // from x import y
				mod = matches[2]
			}
			
			if mod != "" {
				graph.Edges = append(graph.Edges, Edge{
					From: fileID, To: "import:" + mod, Type: "imports",
				})
			}
		}

		if matches := pyFuncRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			exported := !strings.HasPrefix(name, "_")
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeFunction,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: exported,
				Language:   "python",
			}
		}

		if matches := pyClassRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeTypeName,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: !strings.HasPrefix(name, "_"),
				Language:   "python",
			}
		}
	}
}

func (rgs *RepoGraphService) indexRustFile(graph *Graph, relPath, content string) {
	fileID := "file:" + relPath

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			continue
		}

		// Rust 'mod' resolution
		if matches := rsModRe.FindStringSubmatch(line); len(matches) > 1 {
			modName := matches[1]
			target := rgs.resolveRustPath(relPath, modName)
			graph.Edges = append(graph.Edges, Edge{
				From: fileID, To: target, Type: "imports",
			})
		}

		// Rust 'use' resolution
		if matches := rsUseRe.FindStringSubmatch(line); len(matches) > 0 {
			basePath := matches[1]
			target := rgs.resolveRustPath(relPath, basePath)
			graph.Edges = append(graph.Edges, Edge{
				From: fileID, To: target, Type: "imports",
			})
		}

		if matches := rsFuncRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			exported := strings.Contains(line, "pub")
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeFunction,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: exported,
				Language:   "rust",
			}
		}

		if matches := rsStructRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			exported := strings.Contains(line, "pub")
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeTypeName,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: exported,
				Language:   "rust",
			}
		}
	}
}

func (rgs *RepoGraphService) resolveRustPath(currentFile, rustPath string) string {
	if rsAbsoluteImportRe.MatchString(rustPath) {
		return "import:" + rustPath
	}
	dir := filepath.Dir(currentFile)
	if dir == "." {
		dir = ""
	}
	parts := strings.Split(rustPath, "::")
	
	if len(parts) == 0 {
		return "import:" + rustPath
	}

	targetDir := dir
	startIdx := 0

	if parts[0] == "crate" {
		// Find the root of the current package (look for Cargo.toml or just go to monorepo root)
		// For now, in our monorepo structure, we'll try to find 'src' parent or just use root
		targetDir = ""
		if strings.Contains(dir, "src") {
			targetDir = strings.Split(dir, "src")[0]
		}
		startIdx = 1
	} else if parts[0] == "super" {
		targetDir = filepath.Dir(dir)
		if targetDir == "." {
			targetDir = ""
		}
		startIdx = 1
	} else if parts[0] == "self" {
		targetDir = dir
		startIdx = 1
	}

	// Build the path from remaining parts
	remaining := parts[startIdx:]
	if len(remaining) == 0 {
		// e.g. use crate;
		target := filepath.ToSlash(targetDir)
		extensions := []string{"/lib.rs", "/main.rs"}
		for _, ext := range extensions {
			if _, ok := rgs.graph.Nodes["file:"+target+ext]; ok {
				return "file:" + target + ext
			}
		}
		return "import:" + rustPath
	}

	target := filepath.ToSlash(filepath.Clean(filepath.Join(targetDir, filepath.Join(remaining...))))

	extensions := []string{".rs", "/mod.rs", "/lib.rs", "/main.rs"}
	for _, ext := range extensions {
		fullPath := target + ext
		if rgs.graph != nil {
			if _, ok := rgs.graph.Nodes["file:"+fullPath]; ok {
				return "file:" + fullPath
			}
		}
	}

	return "import:" + rustPath
}

// --- Helpers ---

func languageFromExt(ext string) string {
	switch ext {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	default:
		return "unknown"
	}
}

func (rgs *RepoGraphService) countByType(graph *Graph, nodeType NodeType) int {
	count := 0
	for _, node := range graph.Nodes {
		if node.Type == nodeType {
			count++
		}
	}
	return count
}

// Ensure sort import is available
var _ = sort.Ints
