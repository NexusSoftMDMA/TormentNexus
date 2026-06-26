package mcpimpl

import (
	"context"
	"net/http"
)

func HandleHatch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	count, _ :=getInt(args, "count")

	if count <= 0 {
		count = 1
	}

	_, e := http.DefaultClient.Get("https://example.com/hatch")
	if e != nil {
		return err("failed to hatch: " + e.Error())
}

	return success("hatched " + name + " x" + formatInt(int64(count)))
}