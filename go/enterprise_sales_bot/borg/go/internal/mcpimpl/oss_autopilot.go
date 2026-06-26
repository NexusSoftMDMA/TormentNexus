package mcpimpl

import (
	"context"
)

func HandleListBuckets_oss_autopilot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("bucket1, bucket2, bucket3")
}

func HandleGetObject_oss_autopilot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bucket, _ :=getString(args, "bucket")
	key, _ :=getString(args, "key")
	if bucket == "" || key == "" {
		return err("bucket and key are required")
	}
	return ok("object data for " + bucket + "/" + key)
}