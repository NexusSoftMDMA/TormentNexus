package tools

import "context"

func HandleListFioriApps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success(`{"apps": [{"id":"F001","name":"My App"}]}`)
}

func HandleGetFioriApp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "appId")
	if id == "" {
		return err("appId is required")
}

	return success(`{"id":"` + id + `","name":"App Details"}`)
}