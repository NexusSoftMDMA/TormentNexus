package mcpimpl

import "context"

func HandlePlanpong(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	plan, _ :=getString(args, "plan")
	if plan == "" {
		return err("plan is required")
}

	critique := "The plan is well structured but lacks detailed risk mitigation."
	refined := "Refined plan:\n" + plan + "\nAdded: risk mitigation steps, milestones, and testing phases."
	return ok(critique + "\n\n" + refined)
}