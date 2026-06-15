//go:build ignore
// +build ignore

package tools

/**
 * @file serena.go
 * @module go/internal/tools
 *
 * WHAT: Go-native implementation of Serena (Semantic Code Understanding) MCP tools.
 * Implements find_declaration, find_implementations, find_symbol, get_symbols_overview,
 * find_referencing_symbols, rename_symbol, and onboarding.
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type LSSymbol struct {
	Name         string      `json:"name"`
	Kind         string      `json:"kind"`
	RelativePath string      `json:"relative_path,omitempty"`
	Location     *LSLocation `json:"location,omitempty"`
	BodyLocation *LSBodyLoc  `json:"body_location,omitempty"`
	Body         string      `json:"body,omitempty"`
	NamePath     string      `json:"name_path,omitempty"`
	Children     []*LSSymbol `json:"children,omitempty"`

	// internal fields for lookup/indexing
	startLine int
	startCol  int
	endLine   int
	parent    *LSSymbol
}

type LSLocation struct {
	RelativePath string `json:"relativePath,omitempty"`
	Line         int    `json:"line"`      // 0-based index
	Character    int    `json:"character"` // 0-based index
}

type LSBodyLoc struct {
	StartLine int `json:"start_line"` // 0-based
	EndLine   int `json:"end_line"`   // 0-based
}

func (s *LSSymbol) ToDict(depth int, includeBody bool, includeLocation bool, includeRelativePath bool) map[string]interface{} {
	res := map[string]interface{}{}
	res["name"] = s.Name
	res["kind"] = s.Kind
	if includeRelativePath {
		res["relative_path"] = s.RelativePath
	}
	if includeLocation && s.Location != nil {
		loc := map[string]interface{}{
			"line":      s.Location.Line,
			"character": s.Location.Character,
		}
		if includeRelativePath {
			loc["relativePath"] = s.Location.RelativePath
		}
		res["location"] = loc
	}
	if s.BodyLocation != nil {
		res["body_location"] = map[string]interface{}{
			"start_line": s.BodyLocation.StartLine,
			"end_line":   s.BodyLocation.EndLine,
		}
	}
	if includeBody {
		res["body"] = s.Body
	}
	res["name_path"] = s.NamePath

	if depth > 0 && len(s.Children) > 0 {
		var children []interface{}
		for _, child := range s.Children {
			children = append(children, child.ToDict(depth-1, includeBody, includeLocation, false))
		}
		if len(children) > 0 {
			res["children"] = children
		}
	}
	return res
}

func parseGoFile(relativePath string, rootDir string) ([]*LSSymbol, error) {
	fullPath := filepath.Join(rootDir, relativePath)
	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	content := string(contentBytes)

	fset := token.NewFileSet()
	fileAST, err := parser.ParseFile(fset, fullPath, contentBytes, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var symbols []*LSSymbol
	var methods []*LSSymbol

	getSource := func(start, end token.Pos) string {
		sIdx := fset.Position(start).Offset
		eIdx := fset.Position(end).Offset
		if sIdx >= 0 && eIdx <= len(content) && sIdx <= eIdx {
			return content[sIdx:eIdx]
		}
		return ""
	}

	for _, decl := range fileAST.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			pos := fset.Position(d.Pos())
			endPos := fset.Position(d.End())
			selPos := fset.Position(d.Name.Pos())

			sym := &LSSymbol{
				Name:         d.Name.Name,
				Kind:         "Function",
				RelativePath: relativePath,
				Location: &LSLocation{
					RelativePath: relativePath,
					Line:         selPos.Line - 1,
					Character:    selPos.Column - 1,
				},
				BodyLocation: &LSBodyLoc{
					StartLine: pos.Line - 1,
					EndLine:   endPos.Line - 1,
				},
				Body:      getSource(d.Pos(), d.End()),
				startLine: pos.Line - 1,
				startCol:  pos.Column - 1,
				endLine:   endPos.Line - 1,
			}

			if d.Recv != nil && len(d.Recv.List) > 0 {
				sym.Kind = "Method"
				recvType := d.Recv.List[0].Type
				var recvTypeName string
				switch t := recvType.(type) {
				case *ast.Ident:
					recvTypeName = t.Name
				case *ast.StarExpr:
					if ident, ok := t.X.(*ast.Ident); ok {
						recvTypeName = ident.Name
					}
				}
				if recvTypeName != "" {
					sym.NamePath = recvTypeName + "/" + sym.Name
					methods = append(methods, sym)
					continue
				}
			} else {
				sym.NamePath = sym.Name
			}
			symbols = append(symbols, sym)

		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					pos := fset.Position(typeSpec.Pos())
					endPos := fset.Position(typeSpec.End())
					selPos := fset.Position(typeSpec.Name.Pos())

					sym := &LSSymbol{
						Name:         typeSpec.Name.Name,
						Kind:         "Class",
						RelativePath: relativePath,
						Location: &LSLocation{
							RelativePath: relativePath,
							Line:         selPos.Line - 1,
							Character:    selPos.Column - 1,
						},
						BodyLocation: &LSBodyLoc{
							StartLine: pos.Line - 1,
							EndLine:   endPos.Line - 1,
						},
						Body:      getSource(typeSpec.Pos(), typeSpec.End()),
						startLine: pos.Line - 1,
						startCol:  pos.Column - 1,
						endLine:   endPos.Line - 1,
					}
					sym.NamePath = sym.Name

					switch structType := typeSpec.Type.(type) {
					case *ast.StructType:
						sym.Kind = "Struct"
						if structType.Fields != nil {
							for _, field := range structType.Fields.List {
								if len(field.Names) == 0 {
									continue
								}
								fieldPos := fset.Position(field.Pos())
								fieldEndPos := fset.Position(field.End())
								for _, nameIdent := range field.Names {
									fieldSelPos := fset.Position(nameIdent.Pos())
									fieldSym := &LSSymbol{
										Name:         nameIdent.Name,
										Kind:         "Field",
										RelativePath: relativePath,
										Location: &LSLocation{
											RelativePath: relativePath,
											Line:         fieldSelPos.Line - 1,
											Character:    fieldSelPos.Column - 1,
										},
										BodyLocation: &LSBodyLoc{
											StartLine: fieldPos.Line - 1,
											EndLine:   fieldEndPos.Line - 1,
										},
										Body:      getSource(field.Pos(), field.End()),
										startLine: fieldPos.Line - 1,
										startCol:  fieldPos.Column - 1,
										endLine:   fieldEndPos.Line - 1,
										parent:    sym,
										NamePath:  sym.Name + "/" + nameIdent.Name,
									}
									sym.Children = append(sym.Children, fieldSym)
								}
							}
						}
					case *ast.InterfaceType:
						sym.Kind = "Interface"
						if structType.Methods != nil {
							for _, method := range structType.Methods.List {
								if len(method.Names) == 0 {
									continue
								}
								mPos := fset.Position(method.Pos())
								mEndPos := fset.Position(method.End())
								for _, nameIdent := range method.Names {
									mSelPos := fset.Position(nameIdent.Pos())
									mSym := &LSSymbol{
										Name:         nameIdent.Name,
										Kind:         "Method",
										RelativePath: relativePath,
										Location: &LSLocation{
											RelativePath: relativePath,
											Line:         mSelPos.Line - 1,
											Character:    mSelPos.Column - 1,
										},
										BodyLocation: &LSBodyLoc{
											StartLine: mPos.Line - 1,
											EndLine:   mEndPos.Line - 1,
										},
										Body:      getSource(method.Pos(), method.End()),
										startLine: mPos.Line - 1,
										startCol:  mPos.Column - 1,
										endLine:   mEndPos.Line - 1,
										parent:    sym,
										NamePath:  sym.Name + "/" + nameIdent.Name,
									}
									sym.Children = append(sym.Children, mSym)
								}
							}
						}
					}
					symbols = append(symbols, sym)
				}
			} else if d.Tok == token.CONST || d.Tok == token.VAR {
				kind := "Variable"
				if d.Tok == token.CONST {
					kind = "Constant"
				}
				for _, spec := range d.Specs {
					valSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					pos := fset.Position(valSpec.Pos())
					endPos := fset.Position(valSpec.End())
					for _, nameIdent := range valSpec.Names {
						selPos := fset.Position(nameIdent.Pos())
						sym := &LSSymbol{
							Name:         nameIdent.Name,
							Kind:         kind,
							RelativePath: relativePath,
							Location: &LSLocation{
								RelativePath: relativePath,
								Line:         selPos.Line - 1,
								Character:    selPos.Column - 1,
							},
							BodyLocation: &LSBodyLoc{
								StartLine: pos.Line - 1,
								EndLine:   endPos.Line - 1,
							},
							Body:      getSource(valSpec.Pos(), valSpec.End()),
							startLine: pos.Line - 1,
							startCol:  pos.Column - 1,
							endLine:   endPos.Line - 1,
							NamePath:  nameIdent.Name,
						}
						symbols = append(symbols, sym)
					}
				}
			}
		}
	}

	for _, method := range methods {
		parts := strings.Split(method.NamePath, "/")
		if len(parts) == 2 {
			structName := parts[0]
			foundStruct := false
			for _, sym := range symbols {
				if sym.Name == structName && (sym.Kind == "Struct" || sym.Kind == "Class" || sym.Kind == "Interface") {
					method.parent = sym
					sym.Children = append(sym.Children, method)
					foundStruct = true
					break
				}
			}
			if !foundStruct {
				symbols = append(symbols, method)
			}
		} else {
			symbols = append(symbols, method)
		}
	}

	return symbols, nil
}

func parseFallbackFile(relativePath string, rootDir string) ([]*LSSymbol, error) {
	fullPath := filepath.Join(rootDir, relativePath)
	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	content := string(contentBytes)
	lines := strings.Split(content, "\n")

	var symbols []*LSSymbol

	reJSClass := regexp.MustCompile(`^\s*(export\s+)?class\s+([A-Za-z0-9_]+)`)
	reJSFunc := regexp.MustCompile(`^\s*(export\s+)?(async\s+)?function\s+([A-Za-z0-9_]+)`)
	rePyClass := regexp.MustCompile(`^\s*class\s+([A-Za-z0-9_]+)`)
	rePyFunc := regexp.MustCompile(`^\s*def\s+([A-Za-z0-9_]+)`)

	for lineIdx, line := range lines {
		var name, kind string
		var matched bool

		if m := reJSClass.FindStringSubmatch(line); len(m) > 0 {
			name = m[len(m)-1]
			kind = "Class"
			matched = true
		} else if m := reJSFunc.FindStringSubmatch(line); len(m) > 0 {
			name = m[len(m)-1]
			kind = "Function"
			matched = true
		} else if m := rePyClass.FindStringSubmatch(line); len(m) > 0 {
			name = m[0]
			name = strings.TrimSpace(strings.TrimPrefix(name, "class "))
			name = strings.Split(name, "(")[0]
			name = strings.Split(name, ":")[0]
			name = strings.TrimSpace(name)
			kind = "Class"
			matched = true
		} else if m := rePyFunc.FindStringSubmatch(line); len(m) > 0 {
			name = m[0]
			name = strings.TrimSpace(strings.TrimPrefix(name, "def "))
			name = strings.Split(name, "(")[0]
			name = strings.TrimSpace(name)
			kind = "Function"
			matched = true
		}

		if matched && name != "" {
			endLine := lineIdx
			sym := &LSSymbol{
				Name:         name,
				Kind:         kind,
				RelativePath: relativePath,
				Location: &LSLocation{
					RelativePath: relativePath,
					Line:         lineIdx,
					Character:    strings.Index(line, name),
				},
				BodyLocation: &LSBodyLoc{
					StartLine: lineIdx,
					EndLine:   endLine,
				},
				startLine: lineIdx,
				startCol:  0,
				endLine:   endLine,
				NamePath:  name,
			}
			symbols = append(symbols, sym)
		}
	}

	return symbols, nil
}

func GetAllSymbolsInFile(relativePath string, rootDir string) ([]*LSSymbol, error) {
	ext := strings.ToLower(filepath.Ext(relativePath))
	if ext == ".go" {
		return parseGoFile(relativePath, rootDir)
	}
	return parseFallbackFile(relativePath, rootDir)
}

func matchNamePath(namePath string, pattern string, substringMatching bool) bool {
	isAbsolute := strings.HasPrefix(pattern, "/")
	cleanPattern := strings.TrimPrefix(pattern, "/")

	patternParts := strings.Split(cleanPattern, "/")
	nameParts := strings.Split(namePath, "/")

	if isAbsolute {
		if len(patternParts) != len(nameParts) {
			return false
		}
		for i := 0; i < len(patternParts); i++ {
			p := patternParts[i]
			n := nameParts[i]
			if i == len(patternParts)-1 && substringMatching {
				if !strings.Contains(n, p) {
					return false
				}
			} else {
				if n != p {
					return false
				}
			}
		}
		return true
	}

	if len(nameParts) < len(patternParts) {
		return false
	}
	startIdx := len(nameParts) - len(patternParts)
	for i := 0; i < len(patternParts); i++ {
		p := patternParts[i]
		n := nameParts[startIdx+i]
		if i == len(patternParts)-1 && substringMatching {
			if !strings.Contains(n, p) {
				return false
			}
		} else {
			if n != p {
				return false
			}
		}
	}
	return true
}

func findFilesInDir(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == "node_modules" || info.Name() == ".git" || info.Name() == "submodules" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".go" || ext == ".js" || ext == ".ts" || ext == ".py" {
			rel, _ := filepath.Rel(dir, path)
			if rel != "" {
				files = append(files, rel)
			}
		}
		return nil
	})
	return files, err
}

// HandleGetSymbolsOverview gets an overview of the top-level symbols defined in a given file.
func HandleGetSymbolsOverview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	relPath, _ := getString(args, "relative_path", "relativePath", "path")
	if relPath == "" {
		return err("relative_path is required")
	}
	depth := getInt(args, "depth")

	cwd, _ := os.Getwd()
	symbols, errParse := GetAllSymbolsInFile(relPath, cwd)
	if errParse != nil {
		return err(fmt.Sprintf("Failed to parse file: %v", errParse))
	}

	var symbolDicts []map[string]interface{}
	for _, sym := range symbols {
		// Overview filters out low-level symbols like Variables and Fields at root
		if sym.Kind == "Variable" || sym.Kind == "Constant" || sym.Kind == "Field" {
			continue
		}
		symbolDicts = append(symbolDicts, sym.ToDict(depth, false, false, false))
	}

	jsonData, _ := json.MarshalIndent(symbolDicts, "", "  ")
	return ok(string(jsonData))
}

// HandleFindSymbol performs global or local search for matching symbols.
func HandleFindSymbol(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ := getString(args, "name_path_pattern", "pattern")
	if pattern == "" {
		return err("name_path_pattern is required")
	}
	relPath, _ := getString(args, "relative_path", "relativePath")
	depth := getInt(args, "depth")
	includeBody := getBool(args, "include_body")
	substringMatching := getBool(args, "substring_matching")

	cwd, _ := os.Getwd()
	var files []string
	if relPath != "" {
		full := filepath.Join(cwd, relPath)
		info, errStat := os.Stat(full)
		if errStat != nil {
			return err(fmt.Sprintf("Path error: %v", errStat))
		}
		if info.IsDir() {
			subFiles, _ := findFilesInDir(full)
			for _, f := range subFiles {
				files = append(files, filepath.Join(relPath, f))
			}
		} else {
			files = append(files, relPath)
		}
	} else {
		files, _ = findFilesInDir(cwd)
	}

	var results []map[string]interface{}
	for _, file := range files {
		syms, _ := GetAllSymbolsInFile(file, cwd)
		for _, sym := range syms {
			// recursively search in symbol hierarchy
			var checkSym func(*LSSymbol)
			checkSym = func(s *LSSymbol) {
				if matchNamePath(s.NamePath, pattern, substringMatching) {
					results = append(results, s.ToDict(depth, includeBody, true, true))
				}
				for _, child := range s.Children {
					checkSym(child)
				}
			}
			checkSym(sym)
		}
	}

	jsonData, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(jsonData))
}

// HandleFindReferencingSymbols finds references to a symbol across the codebase.
func HandleFindReferencingSymbols(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namePath, _ := getString(args, "name_path")
	relPath, _ := getString(args, "relative_path")
	if namePath == "" || relPath == "" {
		return err("name_path and relative_path are required")
	}

	cwd, _ := os.Getwd()
	// Find the symbol's base name (e.g. function/method name or struct name)
	parts := strings.Split(namePath, "/")
	symbolName := parts[len(parts)-1]

	// Find files containing this symbolName as a word/token
	files, _ := findFilesInDir(cwd)
	var references []map[string]interface{}

	reWord := regexp.MustCompile(`\b` + regexp.QuoteMeta(symbolName) + `\b`)

	for _, file := range files {
		fullPath := filepath.Join(cwd, file)
		data, errRead := os.ReadFile(fullPath)
		if errRead != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		syms, _ := GetAllSymbolsInFile(file, cwd)

		for lineIdx, line := range lines {
			if reWord.MatchString(line) {
				// Don't count the definition itself as a reference
				isDefinition := false
				if file == relPath {
					for _, sym := range syms {
						if sym.Name == symbolName && sym.startLine == lineIdx {
							isDefinition = true
							break
						}
					}
				}
				if isDefinition {
					continue
				}

				// Find enclosing symbol
				var enclosing *LSSymbol
				for _, sym := range syms {
					var checkEnclosing func(*LSSymbol)
					checkEnclosing = func(s *LSSymbol) {
						if lineIdx >= s.startLine && lineIdx <= s.endLine {
							if enclosing == nil || (s.endLine-s.startLine) < (enclosing.endLine-enclosing.startLine) {
								enclosing = s
							}
						}
						for _, child := range s.Children {
							checkEnclosing(child)
						}
					}
					checkEnclosing(sym)
				}

				ref := map[string]interface{}{
					"relative_path":  file,
					"reference_line": lineIdx,
				}
				if enclosing != nil {
					ref["enclosing_symbol"] = enclosing.ToDict(0, false, true, true)
				}
				// Get context lines
				startLine := lineIdx - 1
				if startLine < 0 {
					startLine = 0
				}
				endLine := lineIdx + 1
				if endLine >= len(lines) {
					endLine = len(lines) - 1
				}
				var contextLines []string
				for l := startLine; l <= endLine; l++ {
					contextLines = append(contextLines, lines[l])
				}
				ref["content_around_reference"] = strings.Join(contextLines, "\n")
				references = append(references, ref)
			}
		}
	}

	jsonData, _ := json.MarshalIndent(references, "", "  ")
	return ok(string(jsonData))
}

// HandleFindImplementations finds symbols implementing an interface.
func HandleFindImplementations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namePath, _ := getString(args, "name_path")
	relPath, _ := getString(args, "relative_path")
	if namePath == "" || relPath == "" {
		return err("name_path and relative_path are required")
	}

	cwd, _ := os.Getwd()
	syms, errParse := GetAllSymbolsInFile(relPath, cwd)
	if errParse != nil {
		return err(fmt.Sprintf("Failed to parse file: %v", errParse))
	}

	var targetInterface *LSSymbol
	for _, sym := range syms {
		if sym.NamePath == namePath && sym.Kind == "Interface" {
			targetInterface = sym
			break
		}
	}

	if targetInterface == nil {
		return ok("[]")
	}

	// For a simple implementation checker, we search for struct symbols
	// in Go files that implement method names similar to the interface.
	allFiles, _ := findFilesInDir(cwd)
	var implementations []map[string]interface{}

	for _, file := range allFiles {
		fileSyms, _ := GetAllSymbolsInFile(file, cwd)
		for _, sym := range fileSyms {
			if sym.Kind == "Struct" || sym.Kind == "Class" {
				// Count matching methods
				matchCount := 0
				for _, ifaceMethod := range targetInterface.Children {
					for _, structMethod := range sym.Children {
						if structMethod.Name == ifaceMethod.Name && structMethod.Kind == "Method" {
							matchCount++
							break
						}
					}
				}
				// If struct implements all methods of the interface (or has methods with same names)
				if matchCount > 0 && matchCount == len(targetInterface.Children) {
					implementations = append(implementations, sym.ToDict(0, false, true, true))
				}
			}
		}
	}

	jsonData, _ := json.MarshalIndent(implementations, "", "  ")
	return ok(string(jsonData))
}

// HandleFindDeclaration finds the declaration/definition of a symbol.
func HandleFindDeclaration(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	relPath, _ := getString(args, "relative_path", "relativePath")
	regexStr, _ := getString(args, "regex")
	containingSymbol, _ := getString(args, "containing_symbol_name_path")
	includeBody := getBool(args, "include_body")

	if relPath == "" || regexStr == "" {
		return err("relative_path and regex are required")
	}

	cwd, _ := os.Getwd()
	fullPath := filepath.Join(cwd, relPath)
	contentBytes, errRead := os.ReadFile(fullPath)
	if errRead != nil {
		return err(fmt.Sprintf("Error reading file: %v", errRead))
	}
	content := string(contentBytes)

	re, errReg := regexp.Compile(regexStr)
	if errReg != nil {
		return err(fmt.Sprintf("Invalid regex: %v", errReg))
	}

	// Limit search content if containing symbol is specified
	searchContent := content
	startOffset := 0
	if containingSymbol != "" {
		syms, _ := GetAllSymbolsInFile(relPath, cwd)
		for _, sym := range syms {
			if sym.NamePath == containingSymbol {
				searchContent = sym.Body
				// Find start offset of body
				startOffset = strings.Index(content, sym.Body)
				if startOffset < 0 {
					startOffset = 0
				}
				break
			}
		}
	}

	matches := re.FindStringSubmatchIndex(searchContent)
	if len(matches) < 4 {
		return err("Regex did not match or did not contain a capturing group")
	}

	// Capture group is at index 2, 3
	symStart := matches[2] + startOffset
	symEnd := matches[3] + startOffset
	symbolName := content[symStart:symEnd]

	// Find the symbol declaration matching symbolName
	// First search in local file, then globally
	syms, _ := GetAllSymbolsInFile(relPath, cwd)
	for _, sym := range syms {
		if sym.Name == symbolName {
			return ok(fmt.Sprintf("[%s]", string(mustMarshal(sym.ToDict(0, includeBody, true, true)))))
		}
	}

	// Search globally
	allFiles, _ := findFilesInDir(cwd)
	for _, file := range allFiles {
		if file == relPath {
			continue
		}
		globalSyms, _ := GetAllSymbolsInFile(file, cwd)
		for _, sym := range globalSyms {
			if sym.Name == symbolName {
				return ok(fmt.Sprintf("[%s]", string(mustMarshal(sym.ToDict(0, includeBody, true, true)))))
			}
		}
	}

	return err(fmt.Sprintf("Declaration of symbol '%s' not found", symbolName))
}

// HandleRenameSymbol renames a symbol throughout the codebase.
func HandleRenameSymbol(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namePath, _ := getString(args, "name_path")
	relPath, _ := getString(args, "relative_path")
	newName, _ := getString(args, "new_name")

	if namePath == "" || relPath == "" || newName == "" {
		return err("name_path, relative_path, and new_name are required")
	}

	cwd, _ := os.Getwd()
	parts := strings.Split(namePath, "/")
	symbolName := parts[len(parts)-1]

	// Find references and declaration files
	files, _ := findFilesInDir(cwd)
	reWord := regexp.MustCompile(`\b` + regexp.QuoteMeta(symbolName) + `\b`)

	count := 0
	for _, file := range files {
		fullPath := filepath.Join(cwd, file)
		data, errRead := os.ReadFile(fullPath)
		if errRead != nil {
			continue
		}
		content := string(data)
		if reWord.MatchString(content) {
			newContent := reWord.ReplaceAllString(content, newName)
			errWrite := os.WriteFile(fullPath, []byte(newContent), 0644)
			if errWrite == nil {
				count++
			}
		}
	}

	return ok(fmt.Sprintf("Renamed %d references of symbol '%s' to '%s'.", count, symbolName, newName))
}

// HandleOnboarding performs onboarding identifying the project structure.
func HandleOnboarding(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt := `You are viewing the project for the first time.
Your task is to assemble durable, non-obvious information about the project and write it to memory files that future agents will consult.

Before writing anything, check if memory conventions exist.
Target memory layout:
* mem:core — top-level source map and project-wide invariants.
* mem:tech_stack — language(s), framework(s), build tools, package manager.
* mem:suggested_commands — project commands (dev, test, lint, format).
* mem:conventions — code style, naming, patterns.
* mem:task_completion — verification checklist.`

	return ok(prompt)
}

func mustMarshal(v interface{}) []byte {
	d, _ := json.Marshal(v)
	return d
}
