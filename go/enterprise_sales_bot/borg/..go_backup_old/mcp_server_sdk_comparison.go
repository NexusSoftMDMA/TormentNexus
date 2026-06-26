package tools

import (
	"context"
)

func HandleCompareSdks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sdk1, _ :=getString(args, "sdk1")
	sdk2, _ :=getString(args, "sdk2")
	if sdk1 == "" || sdk2 == "" {
		return err("Both sdk1 and sdk2 are required")
}

	msg := "Comparing " + sdk1 + " with " + sdk2 + ". Check documentation for details."
	return ok(msg)
}

func HandleListSdks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Available SDKs: Python, Go, TypeScript, Rust, Java")
}