//go:build ignore
// +build ignore

package tools

/**
 * @file prompt_library.go
 * @module go/internal/tools
 *
 * WHAT: Prompt Library — SQLite-backed prompt storage and retrieval.
 * Track D: Migrate hardcoded prompts to database.
 *
 * Provides list/get/search for prompts stored in data/prompt_library.db.
 *
 * Tools:
 *  - prompt_list — list prompt names and descriptions (no content)
 *  - prompt_get — get full prompt content by name
 *  - prompt_search — search prompts by keyword
 */

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	promptDBCache     []promptSummary
	promptDBCacheTime time.Time
	promptDBCacheMu   sync.RWMutex
)

type promptSummary struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Tags        string `json:"tags"`
}

func getPromptDBPath() string {
	candidates := []string{
		"data/prompt_library.db",
		"../data/prompt_library.db",
		"prompt_library.db",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "data/prompt_library.db"
}

func openPromptDB() (*sql.DB, error) {
	return sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", getPromptDBPath()))
}

func loadPromptSummaries(ctx context.Context) ([]promptSummary, error) {
	promptDBCacheMu.RLock()
	if time.Since(promptDBCacheTime) < 5*time.Minute && len(promptDBCache) > 0 {
		cache := promptDBCache
		promptDBCacheMu.RUnlock()
		return cache, nil
	}
	promptDBCacheMu.RUnlock()

	promptDBCacheMu.Lock()
	defer promptDBCacheMu.Unlock()

	db, e := openPromptDB()
	if e != nil {
		return nil, fmt.Errorf("db open: %v", e)
	}
	defer db.Close()

	rows, e := db.QueryContext(ctx, "SELECT id, name, description, category, tags FROM prompts ORDER BY name")
	if e != nil {
		return nil, fmt.Errorf("query: %v", e)
	}
	defer rows.Close()

	var out []promptSummary
	for rows.Next() {
		var s promptSummary
		rows.Scan(&s.ID, &s.Name, &s.Description, &s.Category, &s.Tags)
		out = append(out, s)
	}
	promptDBCache = out
	promptDBCacheTime = time.Now()
	return out, nil
}

// HandlePromptList returns all prompt names and descriptions (no content).
func HandlePromptList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	summaries, e := loadPromptSummaries(ctx)
	if e != nil {
		return err(fmt.Sprintf("load failed: %v", e))
	}

	category, _ := getString(args, "category")
	if category != "" {
		var filtered []promptSummary
		for _, s := range summaries {
			if strings.EqualFold(s.Category, category) {
				filtered = append(filtered, s)
			}
		}
		summaries = filtered
	}

	data, _ := json.Marshal(summaries)
	return ok(string(data))
}

// HandlePromptGet returns full prompt content.
func HandlePromptGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ := getString(args, "name", "id")
	if name == "" {
		return err("name or id is required")
	}

	db, e := openPromptDB()
	if e != nil {
		return err(fmt.Sprintf("db open: %v", e))
	}
	defer db.Close()

	var content, description, category string
	e = db.QueryRowContext(ctx,
		"SELECT content, description, category FROM prompts WHERE name=? OR CAST(id AS TEXT)=? LIMIT 1",
		name, name).Scan(&content, &description, &category)
	if e == sql.ErrNoRows {
		return err("prompt not found: " + name)
	}
	if e != nil {
		return err(fmt.Sprintf("query error: %v", e))
	}

	// Increment usage count
	db.ExecContext(ctx, "UPDATE prompts SET usage_count=usage_count+1, updated_at=CURRENT_TIMESTAMP WHERE name=?", name)

	result := fmt.Sprintf("**%s** (%s)\n*%s*\n\n%s", name, category, description, content)
	return ok(result)
}

// HandlePromptSearch searches prompts by keyword.
func HandlePromptSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query is required")
	}

	db, e := openPromptDB()
	if e != nil {
		return err(fmt.Sprintf("db open: %v", e))
	}
	defer db.Close()

	like := "%" + query + "%"
	rows, e := db.QueryContext(ctx,
		"SELECT id, name, description, category, tags FROM prompts WHERE name LIKE ? OR description LIKE ? OR content LIKE ? LIMIT 20",
		like, like, like)
	if e != nil {
		return err(fmt.Sprintf("search error: %v", e))
	}
	defer rows.Close()

	var results []promptSummary
	for rows.Next() {
		var s promptSummary
		rows.Scan(&s.ID, &s.Name, &s.Description, &s.Category, &s.Tags)
		results = append(results, s)
	}

	// Sort by relevance (exact name match first, then description match)
	sort.Slice(results, func(i, j int) bool {
		ni := strings.EqualFold(results[i].Name, query)
		nj := strings.EqualFold(results[j].Name, query)
		if ni != nj {
			return ni
		}
		return results[i].Name < results[j].Name
	})

	data, _ := json.Marshal(results)
	return ok(string(data))
}
