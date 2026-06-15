package tools

import "context"

func HandleGetMuiComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "component_name")
	if name == "" {
		return err("component_name is required")
}

	return ok("Component " + name + " is a MUI component. Documentation: https://mui.com/material-ui/react-" + name + "/")
}

func HandleListMuiComponents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	components := []string{"Button", "TextField", "AppBar", "Card", "Dialog"}
	list := "Available MUI components:\n"
	for _, c := range components {
		list += "- " + c + "\n"
	}
	return success(list)
}