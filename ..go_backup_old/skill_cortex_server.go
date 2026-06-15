package tools

import "context"

func HandleGetSkills(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query != "" {
		return success("Skills matching '" + query + "': skill1, skill2")
}

	return success("All skills: skill1, skill2, skill3")
}

func HandleGetSkill(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("Missing skill id")
}

	return success("Skill " + id + ": description of skill")
}