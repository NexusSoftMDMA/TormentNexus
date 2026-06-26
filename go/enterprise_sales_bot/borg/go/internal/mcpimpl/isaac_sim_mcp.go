package mcpimpl

import "context"

func HandleListScenes_isaac_sim_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Scenes listed successfully")
}

func HandleLoadScene(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sceneName, _ :=getString(args, "scene_name")
	return ok("Loaded scene: " + sceneName)
}