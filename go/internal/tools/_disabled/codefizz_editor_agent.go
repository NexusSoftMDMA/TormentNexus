package tools

import "context"

func HandleCreateMaterial(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "materialName")
	return success("Created material: " + name)
}

func HandleEditBlueprint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "blueprintPath")
	return success("Edited blueprint: " + path)
}// touch 1781132122
