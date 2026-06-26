package ai

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const SwarmPromptPlanner = `You are the Swarm Planner. Your goal is to architect a high-level implementation strategy for the task.
Focus on:
- Structural changes needed.
- Required tools and dependencies.
- Potential implementation pitfalls.
Break the task into logical steps for the Implementer.
Use the tormentnexus__repograph_search and tormentnexus__repograph_find_references tools to understand code structure before planning.`

const SwarmPromptImplementer = `You are the Swarm Implementer. Your goal is to write the actual code and execute the necessary tools.
Focus on:
- Following the provided plan precisely.
- Writing clean, maintainable code.
- Verifying changes as you go.
Use tormentnexus__repograph_find_references to perform impact analysis on any exported symbols you modify.`

const SwarmPromptTester = `You are the Swarm Tester. Your goal is to verify the implementation against the plan and requirements.
Focus on:
- Correctness and performance.
- Edge cases and security vulnerabilities.
- Integration with existing modules.
Use tormentnexus__repograph_find_dependents to identify all files that must be re-tested after these changes.`

const SwarmPromptCritic = `You are the Swarm Critic. Your goal is to evaluate the collective progress of the swarm.
Focus on:
- Has the original goal been met?
- Is the current transcript reaching consensus?
- What is missing or needs refinement?
If the task is complete, start your response with "COMPLETE".`

func getSkillsDBPath() string {
	if _, err := os.Stat(".tormentnexus/skills.db"); err == nil {
		return ".tormentnexus/skills.db"
	}
	if home, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(home, ".tormentnexus", "skills.db")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func GetSwarmPrompt(role string) string {
	dbPath := getSkillsDBPath()
	if dbPath != "" {
		if db, err := sql.Open("sqlite", dbPath); err == nil {
			defer db.Close()
			var content string
			query := "SELECT content FROM skills WHERE skill_id = ? OR (category = 'prompt' AND name LIKE ?)"
			err = db.QueryRow(query, "swarm_prompt_"+role, "%"+role+"%").Scan(&content)
			if err == nil && content != "" {
				if strings.HasPrefix(content, "---") {
					parts := strings.SplitN(content, "---", 3)
					if len(parts) >= 3 {
						return strings.TrimSpace(parts[2])
					}
				}
				return strings.TrimSpace(content)
			}
		}
	}

	switch role {
	case "planner":
		return SwarmPromptPlanner
	case "implementer":
		return SwarmPromptImplementer
	case "tester":
		return SwarmPromptTester
	case "critic":
		return SwarmPromptCritic
	default:
		return "You are a helpful assistant."
	}
}

