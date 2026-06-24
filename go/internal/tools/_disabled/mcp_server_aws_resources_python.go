package tools

import (
	"context"
	"fmt"
)

func HandleListResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	region, _ :=getString(args, "region")
	if region == "" {
		region = "us-east-1"
	}
	msg := fmt.Sprintf("Listing AWS resources in region %s. Supported: S3 buckets, EC2 instances, Lambda functions.", region)
	return success(msg)
}