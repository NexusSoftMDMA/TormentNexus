package tools

import (
	"context"
	"fmt"
	"time"
)

// HandleGenerateSnowflake generates a simple snowflake ID.
func HandleGenerateSnowflake(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	nodeID, _ :=getInt(args, "nodeId")
	if nodeID < 0 || nodeID > 1023 {
		return err("nodeId must be between 0 and 1023")
}

	epoch := int64(1577836800000) // 2020-01-01T00:00:00Z in milliseconds
	now := time.Now().UnixMilli()
	timestamp := now - epoch

	snowflake := (timestamp << 22) | (int64(nodeID) << 12) | (0 & 0xFFF)
	return ok(fmt.Sprintf("%d", snowflake))
}