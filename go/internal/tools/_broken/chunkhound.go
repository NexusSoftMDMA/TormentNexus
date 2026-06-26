package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/memorystore"
)

func HandleSearchSemantic(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query")
	if query == "" {
		return err("query is required")
	}

	limit, okLimit := getInt(args, "limit")
	if !okLimit || limit <= 0 {
		limit = 10
	}

	cwd, _ := os.Getwd()
	results, queryErr := memorystore.Search(cwd, query, limit)
	if queryErr != nil {
		return err("semantic search failed: " + queryErr.Error())
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Semantic Search results for: %q\n\n", query))
	if len(results) == 0 {
		builder.WriteString("No semantic matches found in database.")
	} else {
		for i, res := range results {
			builder.WriteString(fmt.Sprintf("[%d] ID: %s (Type: %s, Source: %s)\n", i+1, res.ID, res.Type, res.Source))
			if res.Title != "" {
				builder.WriteString(fmt.Sprintf("Title: %s\n", res.Title))
			}
			builder.WriteString(fmt.Sprintf("Content: %s\n\n", res.Content))
		}
	}

	return ok(builder.String())
}

func HandleSearchRegex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	patternStr, _ := getString(args, "pattern")
	if patternStr == "" {
		return err("pattern is required")
	}

	dir, _ := getString(args, "dir")
	if dir == "" {
		dir = "."
	}

	rx, compileErr := regexp.Compile(patternStr)
	if compileErr != nil {
		return err("invalid regex: " + compileErr.Error())
	}

	var matches []string
	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == ".tormentnexus" {
				return filepath.SkipDir
			}
			return nil
		}

		// Read files up to 1MB
		if info.Size() > 1024*1024 {
			return nil
		}

		contentBytes, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		content := string(contentBytes)
		if rx.MatchString(content) {
			lines := strings.Split(content, "\n")
			for lineNum, line := range lines {
				if rx.MatchString(line) {
					matches = append(matches, fmt.Sprintf("%s:%d: %s", path, lineNum+1, strings.TrimSpace(line)))
				}
				if len(matches) > 100 {
					return fmt.Errorf("limit exceeded")
				}
			}
		}
		return nil
	})

	if walkErr != nil && walkErr.Error() != "limit exceeded" {
		return err("file walk failed: " + walkErr.Error())
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Regex Search matches for: %q in dir %q\n\n", patternStr, dir))
	if len(matches) == 0 {
		builder.WriteString("No pattern matches found.")
	} else {
		builder.WriteString(strings.Join(matches, "\n"))
	}

	return ok(builder.String())
}

func HandleCodeResearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query")
	if query == "" {
		return err("query is required")
	}

	dir, _ := getString(args, "dir")
	if dir == "" {
		dir = "."
	}

	// We combine regex search and semantic search to simulate deep research
	semanticArgs := map[string]interface{}{"query": query, "limit": 5}
	regexArgs := map[string]interface{}{"pattern": query, "dir": dir}

	semResp, _ := HandleSearchSemantic(ctx, semanticArgs)
	regResp, _ := HandleSearchRegex(ctx, regexArgs)

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("# ChunkHound Code Research Report: %q\n\n", query))
	builder.WriteString("## Semantic Context Matches\n")
	if len(semResp.Content) > 0 {
		builder.WriteString(semResp.Content[0].Text)
	} else {
		builder.WriteString("No semantic matching records found.\n")
	}
	builder.WriteString("\n## Pattern Match References\n")
	if len(regResp.Content) > 0 {
		builder.WriteString(regResp.Content[0].Text)
	} else {
		builder.WriteString("No direct text occurrences found.\n")
	}

	return ok(builder.String())
}
