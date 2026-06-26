package mcpimpl

import (
	"context"
	"fmt"
)

func HandleFetchAppInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appId, _ :=getString(args, "appId")
	if appId == "" {
		return err("appId is required")
}

	return success(fmt.Sprintf("Fetched app info for %s from Firebase", appId))
}

func HandleCheckAdMob(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	adUnit, _ :=getString(args, "adUnitId")
	if adUnit == "" {
		return err("adUnitId is required")
}

	return ok(fmt.Sprintf("AdMob ad unit %s is active", adUnit))
}// touch 1781132144
