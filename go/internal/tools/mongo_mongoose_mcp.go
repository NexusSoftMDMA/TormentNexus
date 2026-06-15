//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
)

var mockDB = map[string][]map[string]interface{}{

func HandleFind(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	col, _ :=getString(args, "collection")
	data, found := mockDB[col]
	if !found {
		return err("collection not found")
	return ok(fmt.Sprintf("Found %d documents in %s", len(data), col))
}

func HandleInsert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	col, _ :=getString(args, "collection")
	// ignore document data for brevity
	mockDB[col] = append(mockDB[col], map[string]interface{}{"inserted": true})
	return success("inserted")
}
}
}