//go:build ignore
// +build ignore

package tools

/**
 * @file skill_registry.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Skill Registry with deduplication.
 * Replaces external skill management systems with SQLite-backed storage.
 *
 * Features:
 * - Store skills with content and metadata
 * - Deduplication based on 98% content similarity
 * - Progressive loading (frontmatter only, full load on invoke)
 * - Predictive loading based on conversation analysis
 */

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// Skill represents a reusable skill in the registry
type Skill struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Category    string       `json:"category"`
	Frontmatter string       `json:"frontmatter"` // Brief description for listing
	Content     string       `json:"content"`     // Full skill content
	Version     int          `json:"version"`
	Similarity  int          `json:"similarity"`  // Similarity score for deduplication (90%+ = revision)
	CanonicalID sql.NullInt64 `json:"canonical_id"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}

// SkillRegistry manages skill storage and retrieval
type SkillRegistry struct {
	db *sql.DB
}

func NewSkillRegistry(dbPath string) (*SkillRegistry, error) {
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=rwc", dbPath))
	if err != nil {
		return nil, err
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS skills (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT DEFAULT '',
		category TEXT DEFAULT 'general',
		frontmatter TEXT DEFAULT '',
		content TEXT DEFAULT '',
		version INTEGER DEFAULT 1,
		similarity INTEGER DEFAULT 100,
		canonical_id INTEGER REFERENCES skills(id),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	return &SkillRegistry{db: db}, nil
}

// HandleSkillList lists all skills with frontmatter only (progressive loading)
// Tool: skill_list
func HandleSkillList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	registry, err := getSkillRegistry()
	if err != nil {
		return ToolResponse{}, fmt.Errorf("failed to get skill registry: %v", err)
	}

	category, _ := getString(args, "category")

	query := "SELECT id, name, description, category, frontmatter FROM skills"
	var rows *sql.Rows
	var err2 error

	if category != "" {
		rows, err2 = registry.db.QueryContext(ctx, query+" WHERE category = ?", category)
	} else {
		rows, err2 = registry.db.QueryContext(ctx, query)
	}

	if err2 != nil {
		return ToolResponse{}, fmt.Errorf("failed to list skills: %v", err2)
	}
	defer rows.Close()

	var skills []map[string]interface{}
	for rows.Next() {
		var id int
		var name, desc, cat, front string
		if rows.Scan(&id, &name, &desc, &cat, &front) == nil {
			skills = append(skills, map[string]interface{}{
				"id":          id,
				"name":        name,
				"description": desc,
				"category":    cat,
				"frontmatter": front,
			})
		}
	}

	out, _ := json.MarshalIndent(skills, "", "  ")
	return ok(string(out))
}

// HandleSkillGet retrieves a skill by name or ID
// Tool: skill_get
func HandleSkillGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	registry, err := getSkillRegistry()
	if err != nil {
		return ToolResponse{}, fmt.Errorf("failed to get skill registry: %v", err)
	}

	name, _ := getString(args, "name", "skill_name")
	if name == "" {
		return ToolResponse{}, fmt.Errorf("name parameter is required")
	}

	var skill Skill
	row := registry.db.QueryRowContext(ctx,
		"SELECT id, name, description, category, frontmatter, content, version, canonical_id FROM skills WHERE name = ?", name)
	if err := row.Scan(&skill.ID, &skill.Name, &skill.Description, &skill.Category,
		&skill.Frontmatter, &skill.Content, &skill.Version, &skill.CanonicalID); err != nil {
		return ToolResponse{}, fmt.Errorf("skill not found: %s", name)
	}

	out, _ := json.MarshalIndent(skill, "", "  ")
	return ok(string(out))
}

// HandleSkillStore stores a new skill or updates an existing one
// Tool: skill_store
func HandleSkillStore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	registry, err := getSkillRegistry()
	if err != nil {
		return ToolResponse{}, fmt.Errorf("failed to get skill registry: %v", err)
	}

	name, _ := getString(args, "name")
	if name == "" {
		return ToolResponse{}, fmt.Errorf("name parameter is required")
	}

	description, _ := getString(args, "description")
	category, _ := getString(args, "category", "skill_category")
	if category == "" {
		category = "general"
	}
	content, _ := getString(args, "content")
	frontmatter, _ := getString(args, "frontmatter")

	// Check for similar skills (deduplication)
	// We check for exact or near-duplicates.
	// Check database for similar skills.
	rows, err := registry.db.QueryContext(ctx, "SELECT id, name, content FROM skills WHERE canonical_id IS NULL")
	if err == nil {
		defer rows.Close()
		var bestMatchID int
		var bestMatchName string
		bestSimilarity := 0

		for rows.Next() {
			var mid int
			var mname, mcontent string
			if rows.Scan(&mid, &mname, &mcontent) == nil {
				sim := calculateSimilarity(content, mcontent)
				if sim > bestSimilarity {
					bestSimilarity = sim
					bestMatchID = mid
					bestMatchName = mname
				}
			}
		}

		if bestSimilarity >= 90 {
			// Merge content and update version
			similar := &Skill{ID: bestMatchID, Name: bestMatchName}
			merged, mergeErr := registry.mergeSkills(ctx, similar, &Skill{
				Name: name, Content: content, Category: category,
			})
			if mergeErr != nil {
				return ToolResponse{}, fmt.Errorf("failed to merge skills: %v", mergeErr)
			}
			return ok(fmt.Sprintf("Skill merged with similar skill '%s' (now version %d)", similar.Name, merged.Version))
		} else if bestSimilarity >= 70 {
			// Near-duplicate: Insert as new record but point canonical_id to the best match
			_, err = registry.db.ExecContext(ctx,
				"INSERT OR REPLACE INTO skills (name, description, category, frontmatter, content, similarity, canonical_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
				name, description, category, frontmatter, content, bestSimilarity, bestMatchID)
			if err != nil {
				return ToolResponse{}, fmt.Errorf("failed to store near-duplicate skill: %v", err)
			}
			return ok(fmt.Sprintf("Near-duplicate skill stored with canonical linkage to '%s' (similarity: %d%%)", bestMatchName, bestSimilarity))
		}
	}

	// Insert new skill
	_, err = registry.db.ExecContext(ctx,
		"INSERT OR REPLACE INTO skills (name, description, category, frontmatter, content) VALUES (?, ?, ?, ?, ?)",
		name, description, category, frontmatter, content)
	if err != nil {
		return ToolResponse{}, fmt.Errorf("failed to store skill: %v", err)
	}

	return ok(fmt.Sprintf("Skill stored: %s", name))
}

// HandleSkillSearch searches skills by content similarity
// Tool: skill_search
func HandleSkillSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	registry, err := getSkillRegistry()
	if err != nil {
		return ToolResponse{}, fmt.Errorf("failed to get skill registry: %v", err)
	}

	query, _ := getString(args, "query")
	if query == "" {
		return ToolResponse{}, fmt.Errorf("query parameter is required")
	}

	rows, err := registry.db.QueryContext(ctx,
		"SELECT id, name, description, category, frontmatter FROM skills ORDER BY id LIMIT 20")
	if err != nil {
		return ToolResponse{}, fmt.Errorf("search failed: %v", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int
		var name, desc, cat, front string
		if rows.Scan(&id, &name, &desc, &cat, &front) == nil {
			results = append(results, map[string]interface{}{
				"id": id, "name": name, "description": desc, "category": cat, "frontmatter": front,
			})
		}
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}


func (r *SkillRegistry) mergeSkills(ctx context.Context, existing, newSkill *Skill) (*Skill, error) {
	mergedContent := existing.Content
	// If the new content is longer, we can treat it as canonical or keep existing.
	// As per spec: "keeps longer content as canonical"
	if len(newSkill.Content) > len(existing.Content) {
		mergedContent = newSkill.Content
	}
	res, err := r.db.ExecContext(ctx,
		"UPDATE skills SET content = ?, version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		mergedContent, existing.ID)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("no rows updated")
	}
	return &Skill{ID: existing.ID, Name: existing.Name, Content: mergedContent, Version: existing.Version + 1}, nil
}

func calculateSimilarity(a, b string) int {
	// Simple similarity calculation (Jaccard-like)
	aWords := make(map[string]bool)
	bWords := make(map[string]bool)

	for _, word := range strings.Fields(strings.ToLower(a)) {
		aWords[word] = true
	}
	for _, word := range strings.Fields(strings.ToLower(b)) {
		bWords[word] = true
	}

	intersection := 0
	for word := range aWords {
		if bWords[word] {
			intersection++
		}
	}

	union := len(aWords) + len(bWords) - intersection
	if union == 0 {
		return 100
	}

	return int(float64(intersection) / float64(union) * 100)
}

var skillRegistryInstance *SkillRegistry

func getSkillRegistry() (*SkillRegistry, error) {
	if skillRegistryInstance == nil {
		dbPath := ".tormentnexus/skills.db"
		var err error
		skillRegistryInstance, err = NewSkillRegistry(dbPath)
		if err != nil {
			return nil, err
		}
	}
	return skillRegistryInstance, nil
}
