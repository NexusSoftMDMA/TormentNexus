package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleGetPatterns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	patterns := []string{"Singleton", "Observer", "Decorator", "Factory", "MVC"}
	data, e := json.Marshal(patterns)
	if e != nil {
		return err("failed to marshal patterns")
}

	return ok(string(data))
}

func HandleGetPattern(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("pattern name is required")
}

	descriptions := map[string]string{
		"Singleton": "Ensures a class has only one instance.",
		"Observer":  "Defines a one-to-many dependency.",
		"Decorator": "Attaches additional responsibilities.",
		"Factory":   "Creates objects without specifying the exact class.",
		"MVC":       "Separates application into Model, View, and Controller.",
	}
	desc, found := descriptions[name]
	if !found {
		return err(fmt.Sprintf("pattern %q not found", name))
}

	return success(desc)
}