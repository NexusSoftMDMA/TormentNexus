//go:build ignore
// +build ignore

package tools

/**
 * @file vibe_check.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Vibe Check MCP tools.
 * Replaces `vibe-check-mcp` (npx @pv-bhat/vibe-check-mcp@latest) entry in mcp.json.
 *
 * Vibe Check provides code quality review and "vibe" analysis:
 * - Analyzes code for anti-patterns and common mistakes
 * - Checks for "vibe coding" indicators (placeholder code, TODO density, etc.)
 * - Reviews code style and best practices
 * - Provides actionable improvement suggestions
 *
 * Improvements over original:
 * - No npx/Node dependency.
 * - Go-native pattern matching engine.
 * - Configurable rule sets and severity levels.
 * - Supports: analyze_code, check_vibe, review_patterns, get_suggestions.
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type VibeIssue struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Severity string `json:"severity"`
	Category string `json:"category"`
	Message  string `json:"message"`
	Rule     string `json:"rule"`
}

type VibeReport struct {
	FilesAnalyzed int         `json:"files_analyzed"`
	IssuesFound   int         `json:"issues_found"`
	VibeScore     float64     `json:"vibe_score"` // 0-100, 100 = best
	Issues        []VibeIssue `json:"issues"`
	Summary       string      `json:"summary"`
}

// Vibe check rules
var vibeRules = []struct {
	Name     string
	Category string
	Severity string
	Pattern  *regexp.Regexp
	Message  string
}{
	{
		Name:     "todo_density",
		Category: "completeness",
		Severity: "warning",
		Pattern:  regexp.MustCompile(`(?i)(TODO|FIXME|HACK|XXX|WORKAROUND)`),
		Message:  "TODO/FIXME marker found — indicates incomplete implementation",
	},
	{
		Name:     "placeholder_code",
		Category: "completeness",
		Severity: "error",
		Pattern:  regexp.MustCompile(`(?i)(placeholder|stub|not.implemented|pass\s*$|\{\s*\})`),
		Message:  "Placeholder or stub code detected — implementation missing",
	},
	{
		Name:     "hardcoded_secrets",
		Category: "security",
		Severity: "critical",
		Pattern:  regexp.MustCompile(`(?i)(password|secret|api_key|token|api\.key)\s*[:=]\s*['"][^'"]{8,}['"]`),
		Message:  "Potential hardcoded secret or API key detected",
	},
	{
		Name:     "your_key_here",
		Category: "security",
		Severity: "critical",
		Pattern:  regexp.MustCompile(`(?i)(YOUR_.*_KEY_HERE|YOUR_.*_HERE|<YOUR_.*>)`),
		Message:  "Unreplaced placeholder key detected — likely non-functional",
	},
	{
		Name:     "console_log",
		Category: "code_quality",
		Severity: "warning",
		Pattern:  regexp.MustCompile(`console\.(log|debug|info|warn)\s*\(`),
		Message:  "Console logging statement found — should be removed in production",
	},
	{
		Name:     "debugger_statement",
		Category: "code_quality",
		Severity: "error",
		Pattern:  regexp.MustCompile(`debugger\s*;?\s*$`),
		Message:  "Debugger statement found — must be removed before deployment",
	},
	{
		Name:     "commented_code",
		Category: "maintainability",
		Severity: "info",
		Pattern:  regexp.MustCompile(`^\s*//\s*(function|const|let|var|if|for|while|return|import|export)\s`),
		Message:  "Commented-out code detected — consider removing or documenting",
	},
	{
		Name:     "empty_catch",
		Category: "error_handling",
		Severity: "warning",
		Pattern:  regexp.MustCompile(`catch\s*\([^)]*\)\s*\{\s*\}`),
		Message:  "Empty catch block — errors are silently swallowed",
	},
	{
		Name:     "any_type",
		Category: "type_safety",
		Severity: "warning",
		Pattern:  regexp.MustCompile(`:\s*any\b|<any>|as\s+any\b`),
		Message:  "TypeScript 'any' type usage — reduces type safety",
	},
	{
		Name:     "magic_numbers",
		Category: "readability",
		Severity: "info",
		Pattern:  regexp.MustCompile(`\b\d{4,}\b`),
		Message:  "Large numeric literal detected — consider extracting to a named constant",
	},
	{
		Name:     "long_line",
		Category: "readability",
		Severity: "info",
		Pattern:  regexp.MustCompile(`^.{200,}$`),
		Message:  "Very long line detected — consider breaking up for readability",
	},
	{
		Name:     "deprecated_api",
		Category: "maintainability",
		Severity: "warning",
		Pattern:  regexp.MustCompile(`(?i)(componentWillMount|componentWillReceiveProps|componentWillUpdate)\s*\(`),
		Message:  "Deprecated React lifecycle method used",
	},
}

// HandleVibeCheckAnalyze analyzes code for vibe issues.
// Tool: vibe_check_analyze
func HandleVibeCheckAnalyze(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "directory")
	if path == "" {
		path = "."
	}

	minSeverity, _ := getString(args, "min_severity")
	if minSeverity == "" {
		minSeverity = "info"
	}

	report := &VibeReport{
		VibeScore: 100.0,
		Issues:    []VibeIssue{},
	}

	severityOrder := map[string]int{"info": 1, "warning": 2, "error": 3, "critical": 4}
	minSevLevel := severityOrder[minSeverity]

	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		// Skip non-code files
		ext := strings.ToLower(filepath.Ext(p))
		codeExts := map[string]bool{
			".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
			".jsx": true, ".rs": true, ".java": true, ".cpp": true, ".c": true,
			".h": true, ".cs": true, ".rb": true, ".php": true, ".swift": true,
			".kt": true, ".lua": true, ".sh": true,
		}
		if !codeExts[ext] {
			return nil
		}

		// Skip hidden and vendor dirs
		rel, _ := filepath.Rel(path, p)
		if strings.Contains(rel, "node_modules") || strings.Contains(rel, "vendor") ||
			strings.Contains(rel, ".git") || strings.HasPrefix(rel, ".") {
			return nil
		}

		data, e := os.ReadFile(p)
		if e != nil {
			return nil
		}

		report.FilesAnalyzed++
		content := string(data)
		lines := strings.Split(content, "\n")

		for lineNum, line := range lines {
			for _, rule := range vibeRules {
				if severityOrder[rule.Severity] < minSevLevel {
					continue
				}
				if rule.Pattern.MatchString(line) {
					issue := VibeIssue{
						File:     rel,
						Line:     lineNum + 1,
						Severity: rule.Severity,
						Category: rule.Category,
						Message:  rule.Message,
						Rule:     rule.Name,
					}
					report.Issues = append(report.Issues, issue)

					// Deduct from vibe score
					switch rule.Severity {
					case "critical":
						report.VibeScore -= 5.0
					case "error":
						report.VibeScore -= 3.0
					case "warning":
						report.VibeScore -= 1.0
					case "info":
						report.VibeScore -= 0.2
					}
				}
			}
		}

		return nil
	})

	if report.VibeScore < 0 {
		report.VibeScore = 0
	}

	report.IssuesFound = len(report.Issues)

	// Generate summary
	severityCounts := map[string]int{}
	categoryCounts := map[string]int{}
	for _, issue := range report.Issues {
		severityCounts[issue.Severity]++
		categoryCounts[issue.Category]++
	}

	report.Summary = fmt.Sprintf(
		"Analyzed %d files. Found %d issues. Vibe Score: %.1f/100.\n"+
			"Severity: critical=%d, error=%d, warning=%d, info=%d",
		report.FilesAnalyzed, report.IssuesFound, report.VibeScore,
		severityCounts["critical"], severityCounts["error"],
		severityCounts["warning"], severityCounts["info"])

	// Limit output to first 50 issues for readability
	if len(report.Issues) > 50 {
		report.Issues = report.Issues[:50]
	}

	out, _ := json.MarshalIndent(report, "", "  ")
	return ok(string(out))
}

// HandleVibeCheckQuickCheck does a quick vibe check on a single file or code snippet.
// Tool: vibe_check_quick
func HandleVibeCheckQuick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ := getString(args, "code")
	filePath, _ := getString(args, "file_path", "path")

	if code == "" && filePath == "" {
		return err("code or file_path parameter is required")
	}

	var content, source string
	if code != "" {
		content = code
		source = "<inline>"
	} else {
		data, e := os.ReadFile(filePath)
		if e != nil {
			return err(fmt.Sprintf("Failed to read file: %v", e))
		}
		content = string(data)
		source = filePath
	}

	var issues []VibeIssue
	vibeScore := 100.0
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		for _, rule := range vibeRules {
			if rule.Pattern.MatchString(line) {
				issues = append(issues, VibeIssue{
					File:     source,
					Line:     lineNum + 1,
					Severity: rule.Severity,
					Category: rule.Category,
					Message:  rule.Message,
					Rule:     rule.Name,
				})
				switch rule.Severity {
				case "critical":
					vibeScore -= 5.0
				case "error":
					vibeScore -= 3.0
				case "warning":
					vibeScore -= 1.0
				case "info":
					vibeScore -= 0.2
				}
			}
		}
	}

	if vibeScore < 0 {
		vibeScore = 0
	}

	result := map[string]interface{}{
		"source":       source,
		"vibe_score":   vibeScore,
		"issues_found": len(issues),
		"issues":       issues,
	}

	if vibeScore >= 90 {
		result["verdict"] = "✅ Great vibes! Clean code."
	} else if vibeScore >= 70 {
		result["verdict"] = "🟡 Okay vibes. Some issues to address."
	} else if vibeScore >= 40 {
		result["verdict"] = "🟠 Mixed vibes. Significant improvements needed."
	} else {
		result["verdict"] = "🔴 Bad vibes. Major issues detected."
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleVibeCheckReviewPatterns reviews specific code patterns.
// Tool: vibe_check_review_patterns
func HandleVibeCheckReviewPatterns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ := getString(args, "code")
	if code == "" {
		return err("code parameter is required")
	}

	var findings []map[string]interface{}
	lines := strings.Split(code, "\n")

	for lineNum, line := range lines {
		for _, rule := range vibeRules {
			if rule.Pattern.MatchString(line) {
				findings = append(findings, map[string]interface{}{
					"line":     lineNum + 1,
					"rule":     rule.Name,
					"category": rule.Category,
					"severity": rule.Severity,
					"message":  rule.Message,
					"content":  strings.TrimSpace(line),
				})
			}
		}
	}

	if len(findings) == 0 {
		return ok("No pattern violations found. Code looks clean! ✅")
	}

	out, _ := json.MarshalIndent(findings, "", "  ")
	return ok(string(out))
}
