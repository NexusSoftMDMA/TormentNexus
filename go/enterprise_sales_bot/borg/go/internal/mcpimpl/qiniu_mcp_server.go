package mcpimpl

import "context"

func HandleListBuckets_qiniu_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bucket, _ :=getString(args, "bucket")
	return ok("Bucket listed: " + bucket)
}