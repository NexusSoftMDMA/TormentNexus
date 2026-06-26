package tools

import (
	"context"
)

func HandleGenerateComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	component := "// Component: " + name + "\n@Component\nexport class " + name + " {}"
	return ok(component)
}

func HandleGenerateService(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	service := "// Service: " + name + "\n@Injectable()\nexport class " + name + "Service {}"
	return ok(service)
}