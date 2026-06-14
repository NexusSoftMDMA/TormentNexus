package tools

import (
	"context"
	"fmt"
)

func HandleListBuckets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	buckets := []string{"my-bucket", "another-bucket"}
	msg := fmt.Sprintf("Buckets: %v", buckets)
	return ok(msg)
}