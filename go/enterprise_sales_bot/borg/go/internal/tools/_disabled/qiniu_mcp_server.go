package tools

import "context"

func HandleListBuckets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bucket, _ :=getString(args, "bucket")
	return ok("Bucket listed: " + bucket)
}