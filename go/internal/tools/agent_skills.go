//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
)

func HandleListSkills(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	skills := []string{"skill1", "skill2", "skill3"}
	return ok(fmt.Sprintf("Available skills: %v", skills))
}

func HandleGetSkillDescription(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "skill_name")
	if name == "" {
		return err("skill_name is required")
}

	descriptions := map[string]string{
		"skill1": "Description for skill1",
		"skill2": "Description for skill2",
	}
	desc, found := descriptions[name]
	if !found {
		return err("skill not found: " + name)
}

	return ok(desc)
}
