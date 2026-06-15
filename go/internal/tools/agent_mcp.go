//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := "Hello, " + name + "!"
	return ok(msg)
}

func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	result := a + b
	data, e := json.Marshal(map[string]int{"sum": result})
	if e != nil {
		return err("failed to marshal result")
}

	return success(string(data))
}
