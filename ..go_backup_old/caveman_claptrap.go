package tools

import "context"

func HandleClaptrap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Caveman"
	}
	return success("Hello from " + name + " Claptrap!")
}