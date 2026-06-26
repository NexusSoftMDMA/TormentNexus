package tools

import "context"

func HandleListAdbpgInstances(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	region, _ :=getString(args, "region")
	if region == "" {
		return err("region is required")
}

	return ok("List of ADB PG instances in region " + region)
}

func HandleDescribeAdbpgInstance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instanceId, _ :=getString(args, "instanceId")
	if instanceId == "" {
		return err("instanceId is required")
}

	return ok("Details for instance " + instanceId)
}