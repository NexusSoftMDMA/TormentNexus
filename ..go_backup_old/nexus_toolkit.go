package tools

import "context"

func HandleOrchestrate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name must not be empty")
}

	return ok("orchestrate completed for " + name)
}

func HandleResearchCatalogReview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	catalog, _ :=getString(args, "catalog")
	if catalog == "" {
		return err("catalog must not be empty")
}

	action, _ :=getString(args, "action")
	if action == "" {
		action = "review"
	}
	return ok("research catalog review " + action + " for " + catalog)
}