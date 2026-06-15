package tools

import (
	"context"
	"encoding/json"
)

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func HandleListSkills(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	skills := []Skill{
		{Name: "generate", Description: "Generate content"},
		{Name: "translate", Description: "Translate text"},
	}
	data, _ := json.Marshal(skills)
	return ok(string(data))
}

func HandleGetSkill(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Skill: " + name)
}