package mcpimpl

import (
	"context"
	"fmt"
	"strings"
	"unicode"
)

func HandleGenerateTOC(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	md, _ :=getString(args, "markdown")
	if md == "" {
		return ok("No content to generate TOC.")
}

	lines := strings.Split(md, "\n")
	var toc []string
	for _, line := range lines {
		trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
		if trimmed == "" {
			continue
		}
		if trimmed[0] != '#' {
			continue
		}
		level := 0
		for _, c := range trimmed {
			if c == '#' {
				level++
			} else {
				break
			}
		}
		title := strings.TrimSpace(trimmed[level:])
		indent := strings.Repeat("  ", level-1)
		anchor := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
		toc = append(toc, fmt.Sprintf("%s- [%s](#%s)", indent, title, anchor))

	if len(toc) == 0 {
		return ok("No headings found.")
}

	return ok(strings.Join(toc, "\n"))
}

}

func HandleLintMarkdown(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	md, _ :=getString(args, "markdown")
	issues := []string{}

	lines := strings.Split(md, "\n")
	consecutiveBlank := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			consecutiveBlank++
			if consecutiveBlank > 1 {
				issues = append(issues, fmt.Sprintf("Line %d: multiple consecutive blank lines", i+1))

		} else {
			consecutiveBlank = 0
		}
		if strings.HasSuffix(line, " ") && len(line) > 0 {
			issues = append(issues, fmt.Sprintf("Line %d: trailing whitespace", i+1))

	}
	if len(issues) == 0 {
		return ok("No issues found.")
}

	return ok(strings.Join(issues, "\n"))
}
}
}