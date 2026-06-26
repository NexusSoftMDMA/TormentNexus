package mcpimpl

import "context"

func HandleGetBurger(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Burger: " + name)
}

func HandleAddTopping(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	topping, _ :=getString(args, "topping")
	if name == "" || topping == "" {
		return err("name and topping are required")
}

	return success("Added " + topping + " to " + name)
}