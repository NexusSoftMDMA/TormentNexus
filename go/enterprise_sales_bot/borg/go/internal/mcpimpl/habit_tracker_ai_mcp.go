package mcpimpl

import "context"

func HandleCreateHabit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success("Habit '" + name + "' created successfully")
}

func HandleLogCompletion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success("Completion logged for habit '" + name + "'")
}