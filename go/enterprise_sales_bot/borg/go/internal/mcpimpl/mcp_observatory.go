package mcpimpl

import "context"

func HandleObservatoryInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Observatory " + name + ": Located at coordinates 28.5° N, 80.0° W. Altitude: 2,400 m. Operational since 1970.")
}