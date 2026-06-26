package repomap

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var defaultIgnoreDirs = map[string]struct{}{
	".git":         {},
	".hg":          {},
	".svn":         {},
	"node_modules": {},
	"dist":         {},
	"build":        {},
	"target":       {},
	".next":        {},
	"coverage":     {},
}

var sourceExtensions = map[string]struct{}{
	".go":   {},
	".ts":   {},
	".tsx":  {},
	".js":   {},
	".jsx":  {},
	".py":   {},
	".rs":   {},
	".java": {},
	".c":    {},
	".cc":   {},
	".cpp":  {},
	".h":    {},
	".hpp":  {},
	".cs":   {},
	".rb":   {},
	".php":  {},
}

var symbolPatterns = []struct {
	re   *regexp.Regexp
	kind string
}{
	{regexp.MustCompile(`(?m)^\s*func\s+(?:\([^)]*\)\s*)?([A-Za-z_][A-Za-z0-9_]*)\s*\(`), "func"},
	{regexp.MustCompile(`(?m)^\s*type\s+([A-Za-z_][A-Za-z0-9_]*)\s+struct\b`), "struct"},
	{regexp.MustCompile(`(?m)^\s*type\s+([A-Za-z_][A-Za-z0-9_]*)\s+interface\b`), "interface"},
	{regexp.MustCompile(`(?m)^\s*type\s+([A-Za-z_][A-Za-z0-9_]*)\b`), "type"},
	{regexp.MustCompile(`(?m)^\s*class\s+([A-Za-z_][A-Za-z0-9_]*)\b`), "class"},
	{regexp.MustCompile(`(?m)^\s*interface\s+([A-Za-z_][A-Za-z0-9_]*)\b`), "interface"},
	{regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:async\s+)?function\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`), "func"},
	{regexp.MustCompile(`(?m)^\s*(?:export\s+)?const\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(?:async\s*)?\(`), "const"},
	{regexp.MustCompile(`(?m)^\s*def\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`), "func"},
}

var identifierPattern = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]{2,}`)

type Symbol struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Line int    `json:"line"`
}

type Entry struct {
	Path    string   `json:"path"`
	Score   int      `json:"score"`
	Symbols []Symbol `json:"symbols,omitempty"`
}

type Options struct {
	BaseDir         string   `json:"baseDir"`
	MentionedFiles  []string `json:"mentionedFiles,omitempty"`
	MentionedIdents []string `json:"mentionedIdents,omitempty"`
	MaxFiles        int      `json:"maxFiles,omitempty"`
	IncludeTests    bool     `json:"includeTests,omitempty"`
}

type Result struct {
	BaseDir string  `json:"baseDir"`
	Entries []Entry `json:"entries"`
	Map     string  `json:"map"`
}

type fileData struct {
	Entry       Entry
	Identifiers map[string]int
}

func Generate(options Options) (Result, error) {
	baseDir := options.BaseDir
	if strings.TrimSpace(baseDir) == "" {
		baseDir = "."
	}
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return Result{}, fmt.Errorf("resolve base dir: %w", err)
	}
	mentionedFiles := normalizeSet(options.MentionedFiles)
	mentionedIdents := normalizeSet(options.MentionedIdents)

	files := make([]fileData, 0, 64)
	err = filepath.Walk(absBase, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := info.Name()
		if info.IsDir() {
			if _, skip := defaultIgnoreDirs[name]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		if !isSourceFile(name, options.IncludeTests) {
			return nil
		}
		rel, err := filepath.Rel(absBase, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(contentBytes)
		symbols := extractSymbols(content)
		files = append(files, fileData{
			Entry: Entry{
				Path:    rel,
				Symbols: symbols,
			},
			Identifiers: extractIdentifiers(content),
		})
		return nil
	})
	if err != nil {
		return Result{}, fmt.Errorf("walk repo: %w", err)
	}

	scores := rankFiles(files, mentionedFiles, mentionedIdents)
	entries := make([]Entry, 0, len(files))
	for _, file := range files {
		entry := file.Entry
		entry.Score = int(math.Round(scores[file.Entry.Path] * 100))
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Score == entries[j].Score {
			return entries[i].Path < entries[j].Path
		}
		return entries[i].Score > entries[j].Score
	})
	if options.MaxFiles > 0 && len(entries) > options.MaxFiles {
		entries = entries[:options.MaxFiles]
	}
	result := Result{BaseDir: absBase, Entries: entries}
	result.Map = renderMap(entries)
	return result, nil
}

func rankFiles(files []fileData, mentionedFiles, mentionedIdents map[string]struct{}) map[string]float64 {
	baseScores := make(map[string]float64, len(files))
	definitions := make(map[string][]string)
	for _, file := range files {
		baseScores[file.Entry.Path] = float64(scoreEntry(file.Entry.Path, file.Entry.Symbols, mentionedFiles, mentionedIdents))
		for _, symbol := range file.Entry.Symbols {
			definitions[strings.ToLower(symbol.Name)] = append(definitions[strings.ToLower(symbol.Name)], file.Entry.Path)
		}
	}

	edges := make(map[string]map[string]float64, len(files))
	for _, file := range files {
		for ident, count := range file.Identifiers {
			for _, dst := range definitions[ident] {
				if dst == file.Entry.Path {
					continue
				}
				if edges[file.Entry.Path] == nil {
					edges[file.Entry.Path] = map[string]float64{}
				}
				edges[file.Entry.Path][dst] += math.Sqrt(float64(count))
			}
		}
	}

	scores := make(map[string]float64, len(baseScores))
	for path, score := range baseScores {
		scores[path] = score
	}
	const damping = 0.85
	for range 6 {
		next := make(map[string]float64, len(baseScores))
		for path, score := range baseScores {
			next[path] = score
		}
		for src, outgoing := range edges {
			total := 0.0
			for _, weight := range outgoing {
				total += weight
			}
			if total == 0 {
				continue
			}
			for dst, weight := range outgoing {
				next[dst] += damping * scores[src] * (weight / total)
			}
		}
		scores = next
	}
	return scores
}

func renderMap(entries []Entry) string {
	var b strings.Builder
	b.WriteString("<repo_map>\n")
	for _, entry := range entries {
		b.WriteString(entry.Path)
		b.WriteByte('\n')
		for _, symbol := range entry.Symbols {
			b.WriteString(fmt.Sprintf("  %s %s:%d\n", symbol.Kind, symbol.Name, symbol.Line))
		}
	}
	b.WriteString("</repo_map>")
	return b.String()
}

func normalizeSet(values []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out[strings.ToLower(filepath.ToSlash(trimmed))] = struct{}{}
	}
	return out
}

func isSourceFile(name string, includeTests bool) bool {
	ext := strings.ToLower(filepath.Ext(name))
	if _, ok := sourceExtensions[ext]; !ok {
		return false
	}
	if includeTests {
		return true
	}
	lower := strings.ToLower(name)
	return !strings.Contains(lower, "_test.") && !strings.Contains(lower, ".test.") && !strings.Contains(lower, ".spec.")
}

func extractSymbols(content string) []Symbol {
	lines := strings.Split(content, "\n")
	seen := map[string]struct{}{}
	out := make([]Symbol, 0, 8)
	for lineIndex, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		for _, pattern := range symbolPatterns {
			match := pattern.re.FindStringSubmatch(line)
			if len(match) < 2 {
				continue
			}
			name := match[1]
			key := fmt.Sprintf("%s:%s:%d", pattern.kind, name, lineIndex+1)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, Symbol{Name: name, Kind: pattern.kind, Line: lineIndex + 1})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Line == out[j].Line {
			if out[i].Kind == out[j].Kind {
				return out[i].Name < out[j].Name
			}
			return out[i].Kind < out[j].Kind
		}
		return out[i].Line < out[j].Line
	})
	return out
}

func extractIdentifiers(content string) map[string]int {
	matches := identifierPattern.FindAllString(content, -1)
	out := make(map[string]int, len(matches))
	for _, match := range matches {
		out[strings.ToLower(match)]++
	}
	return out
}

func scoreEntry(path string, symbols []Symbol, mentionedFiles, mentionedIdents map[string]struct{}) int {
	score := 1
	lowerPath := strings.ToLower(path)
	if _, ok := mentionedFiles[lowerPath]; ok {
		score += 1000
	}
	for _, symbol := range symbols {
		if _, ok := mentionedIdents[strings.ToLower(symbol.Name)]; ok {
			score += 100
		}
	}
	for component := range splitComponents(path) {
		if _, ok := mentionedIdents[component]; ok {
			score += 25
		}
	}
	score += min(len(symbols), 20)
	return score
}

func splitComponents(path string) map[string]struct{} {
	out := map[string]struct{}{}
	trimmed := strings.TrimSpace(filepath.ToSlash(path))
	for _, part := range strings.Split(trimmed, "/") {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		out[part] = struct{}{}
		base := strings.TrimSuffix(part, filepath.Ext(part))
		if base != "" {
			out[base] = struct{}{}
		}
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
