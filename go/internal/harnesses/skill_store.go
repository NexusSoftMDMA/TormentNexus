package harnesses

/**
 * @file skill_store.go
 * @module go/internal/harnesses
 *
 * WHAT: Go-native implementation of the Skill Store.
 * Manages discovery and persistence of Borg runbooks (.md files).
 *
 * WHY: Total Autonomy — The Go sidecar must be capable of loading and 
 * creating skills without relying on the Node control plane.
 */

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/borghq/borg-go/internal/skillregistry"
)

type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Path        string `json:"path"`
}

type SkillStore struct {
	baseDir  string
	registry *skillregistry.SkillRegistry
}

func NewSkillStore(mainConfigDir string, registry *skillregistry.SkillRegistry) *SkillStore {
	path := filepath.Join(mainConfigDir, "skills")
	_ = os.MkdirAll(path, 0755)
	return &SkillStore{
		baseDir:  path,
		registry: registry,
	}
}

func (s *SkillStore) SaveSkill(id, name, description, content string) error {
	skillDir := filepath.Join(s.baseDir, id)
	_ = os.MkdirAll(skillDir, 0755)

	filePath := filepath.Join(skillDir, "SKILL.md")
	
	// Format with frontmatter
	fullContent := fmt.Sprintf(`---
name: %s
description: %s
---

%s`, name, description, content)

	err := os.WriteFile(filePath, []byte(fullContent), 0644)
	if err != nil {
		return err
	}

	// Update registry
	if s.registry != nil {
		_ = s.registry.Register(skillregistry.SkillInfo{
			ID:          id,
			Name:        name,
			Description: description,
			Content:     content,
			Path:        filePath,
		})
	}

	return nil
}

func (s *SkillStore) ListSkills() ([]string, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, entry := range entries {
		if entry.IsDir() {
			ids = append(ids, entry.Name())
		}
	}
	return ids, nil
}

func (s *SkillStore) GetSkill(id string) (*Skill, error) {
	filePath := filepath.Join(s.baseDir, id, "SKILL.md")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	// Basic parsing of frontmatter
	name := id
	desc := ""
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			// Extract metadata
			lines := strings.Split(parts[1], "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "name:") {
					name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				}
				if strings.HasPrefix(line, "description:") {
					desc = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				}
			}
			content = parts[2]
		}
	}

	return &Skill{
		ID:          id,
		Name:        name,
		Description: desc,
		Content:     content,
		Path:        filePath,
	}, nil
}

