package tools

import (
	"context"
	"fmt"
)

func HandleMem0(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	val, _ :=getString(args, "key")
	if val == "" {
		return err("missing key")
	}
	return ok("mem0 tool called with key: " + val)
}

func HandleMem1(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	val, _ :=getInt(args, "value")
	if val == 0 {
		return err("missing value")
	}
	return ok(fmt.Sprintf("mem1 tool called with value: %d", val))
}

func HandleMem2(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	val, _ :=getBool(args, "flag")
	if !val {
		return err("missing flag")
	}
	return ok("mem2 tool called with flag: true")
}