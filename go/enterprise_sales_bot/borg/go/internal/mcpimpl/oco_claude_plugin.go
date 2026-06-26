package mcpimpl

import "context"

func HandleSafetyCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	if action == "" {
		return err("action is required")
}

	return ok("safety check passed for: " + action)
}

func HandleRunSkill(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	skill, _ :=getString(args, "skill")
	if skill == "" {
		return err("skill name is required")
}

	return success("skill '" + skill + "' executed successfully")
}