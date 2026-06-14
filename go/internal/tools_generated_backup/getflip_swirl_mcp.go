package tools

import "context"

func HandleSwirlComponents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return success(`{"components":["Button","Card","Dialog"]}`)
}

	return success(`{"components":["` + name + `"]}`)
}

func HandleSwirlComponentDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success(`{"component":"` + name + `","description":"A Swirl component"}`)
}