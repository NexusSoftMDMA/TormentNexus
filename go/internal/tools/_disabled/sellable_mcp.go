package tools

import "context"

func HandleCampaignStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	campaignID, _ :=getString(args, "campaignId")
	if campaignID == "" {
		return err("campaignId is required")
}

	return ok("Campaign " + campaignID + " status: active")
}

func HandleCampaignAction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	campaignID, _ :=getString(args, "campaignId")
	action, _ :=getString(args, "action")
	if campaignID == "" || action == "" {
		return err("campaignId and action are required")
}

	return success("Action " + action + " on campaign " + campaignID + " executed")
}