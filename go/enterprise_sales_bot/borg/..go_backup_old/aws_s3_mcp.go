package tools

import (
	"context"
)

func HandleListBuckets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Buckets: [\"my-bucket\"]")
}

func HandleListObjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bucket, _ :=getString(args, "bucket")
	if bucket == "" {
		return err("bucket is required")
}

	return ok("Objects in " + bucket + ": [\"file1.txt\"]")
}