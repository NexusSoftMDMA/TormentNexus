//go:build ignore
// +build ignore

package tools

import (
	"context"
	"strings"
)

func HandleLoadSkill(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "skill_name")
	if name == "" {
		return err("skill_name is required")
}

	return success("Loaded skill: " + name)
}

func HandleListSkills(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	skills := []string{"skill1", "skill2"}
	return ok("Available skills: " + strings.Join(skills, ", "))
}
