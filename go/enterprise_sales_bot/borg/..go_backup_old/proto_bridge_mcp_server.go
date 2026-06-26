package tools

import (
	"context"
)

func HandleMigrationContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		key = "default"
	}
	return success("ProtoBridge migration context loaded for " + key)
}

func HandleYouFiFlutterConventions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		target = "YouFi"
	}
	return success("YouFi Flutter target conventions applied for " + target)
}