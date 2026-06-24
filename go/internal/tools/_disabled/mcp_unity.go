package tools

import (
	"context"
	"fmt"
)

func HandleGetUnityVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Unity version: 2022.3.0f1")
}

func HandleGetSceneList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	scene, _ :=getString(args, "scene")
	if scene == "" {
		return err("scene name is required")
}

	return ok(fmt.Sprintf("Scene '%s' loaded", scene))
}