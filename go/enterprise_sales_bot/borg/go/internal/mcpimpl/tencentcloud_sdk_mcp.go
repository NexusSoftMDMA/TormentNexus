package mcpimpl

import (
	"context"
	"encoding/json"
)

// HandleListRegions returns available Tencent Cloud regions
func HandleListRegions_tencentcloud_sdk_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	regions := []string{"ap-guangzhou", "ap-shanghai", "ap-beijing"}
	data, e := json.Marshal(regions)
	if e != nil {
		return err("failed to marshal regions")
}

	return success(string(data))
}

// HandleDescribeZones returns zones for a given region
func HandleDescribeZones(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	region, _ :=getString(args, "region")
	if region == "" {
		region = "ap-guangzhou"
	}
	zones := []string{region + "-a", region + "-b"}
	data, e := json.Marshal(zones)
	if e != nil {
		return err("failed to marshal zones")
}

	return success(string(data))
}