package tools

import (
	"context"
	"fmt"
)

func HandleGetStorageCostPerGB(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	provider, _ :=getString(args, "provider")
	var costPerGB float64
	switch provider {
	case "aws":
		costPerGB = 0.023
	case "azure":
		costPerGB = 0.02
	case "gcp":
		costPerGB = 0.026
	default:
		return err("unknown provider")
}

	return ok(fmt.Sprintf("Storage cost per GB for %s: $%.3f", provider, costPerGB))
}

func HandleCalculateStorageCost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	provider, _ :=getString(args, "provider")
	sizeGB, _ :=getInt(args, "size_gb")
	var costPerGB float64
	switch provider {
	case "aws":
		costPerGB = 0.023
	case "azure":
		costPerGB = 0.02
	case "gcp":
		costPerGB = 0.026
	default:
		return err("unknown provider")
}

	totalCost := float64(sizeGB) * costPerGB
	return ok(fmt.Sprintf("Total storage cost for %d GB on %s: $%.2f", sizeGB, provider, totalCost))
}