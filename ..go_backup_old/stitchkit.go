package tools

import "context"

func HandleCreateContract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	// contrat creation logic would go here
	return success("contract created: " + name)
}

func HandleGetContract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("contract details for " + name)
}