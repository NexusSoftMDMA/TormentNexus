package tools

import "context"

func HandleListDashboards(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("dashboards: [{'id':'1','name':'Sales Dashboard'}]")
}

func HandleGetDashboard(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "dashboard_id")
	if id == "" {
		return err("missing dashboard_id")
}

	return success("dashboard: {'id':'" + id + "','name':'Sample Dashboard'}")
}