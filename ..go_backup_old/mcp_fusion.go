package tools

import (
	"context"
	"fmt"
)

func HandleFusionInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Fusion"
	}
	return ok(name + " is a process where atomic nuclei combine to release energy.")
}

func HandleFusionCalc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	m1, _ :=getInt(args, "mass1")
	m2, _ :=getInt(args, "mass2")
	if m1 <= 0 || m2 <= 0 {
		return err("Both masses must be positive integers.")
}

	result := m1 * m2
	return ok(fmt.Sprintf("Fusion energy output: %d units.", result))
}