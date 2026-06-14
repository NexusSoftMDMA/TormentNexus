package tools

import "context"

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Stitch Ai User"
	}
	return ok("Hello, " + name + "!")
}