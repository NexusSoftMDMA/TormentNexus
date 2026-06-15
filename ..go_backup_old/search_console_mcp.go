package tools

import "context"

func HandleSearchAnalytics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	siteUrl, _ :=getString(args, "siteUrl")
	startDate, _ :=getString(args, "startDate")
	endDate, _ :=getString(args, "endDate")
	dimensions, _ :=getString(args, "dimensions")
	return ok("Search Analytics: site=" + siteUrl + " from " + startDate + " to " + endDate + " dimensions=" + dimensions)
}

func HandleListSites(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Sites: [https://example.com, https://test.com]")
}