//go:build ignore
// +build ignore

package tools

import (
	"context"
)

func HandleListEmployees(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	e, _ :=getString(args, "query")
	if e != "" {
		return success(`[{"id":1,"name":"John Doe","department":"Engineering"},{"id":2,"name":"Jane Smith","department":"Marketing"}]`)
}

	return success(`[{"id":1,"name":"John Doe","department":"Engineering"},{"id":2,"name":"Jane Smith","department":"Marketing"}]`)
}

func HandleGetEmployee(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	if id == 0 {
		return err("missing employee id")
}

	return success(`{"id":` + string(rune(id)) + `,"name":"Employee ` + string(rune(id)) + `","department":"General"}`)
}
