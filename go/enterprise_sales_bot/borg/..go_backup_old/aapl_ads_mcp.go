package tools

import "context"

func HandleGetAccountInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accountID, _ :=getString(args, "accountId")
	if accountID == "" {
		return err("accountId is required")
}

	return ok("Retrieved account info for " + accountID)
}

func HandleListCampaigns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = getInt(args, "limit")
	_ = getBool(args, "activeOnly")
	return success("Campaigns listed successfully")
}