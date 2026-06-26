package mcpimpl

import (
	"context"
)

// HandleCreateCampaign creates a new Google Ads campaign (mock).
func HandleCreateCampaign(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	budget, _ :=getInt(args, "budget")
	status, _ :=getString(args, "status")
	if name == "" || budget <= 0 {
		return err("invalid campaign parameters")
}

	// Simulate creation
	_ = status
	return ok("campaign created: " + name)
}

// HandleGetCampaign retrieves a campaign by ID (mock).
func HandleGetCampaign_google_ads_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "campaign_id")
	if id == "" {
		return err("campaign_id is required")
}

	// Simulate retrieval
	campaign := map[string]interface{}{
		"id":     id,
		"name":   "Mock Campaign",
		"budget": 10000,
		"status": "ENABLED",
	}
	return success(campaign)
}