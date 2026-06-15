//go:build ignore
// +build ignore

package tools

import (
	"context"
	"time"
)

func HandleGetTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	format, _ :=getString(args, "format")
	if format == "" {
		return ok(time.Now().Format(time.RFC3339))
}

	t := time.Now()
	switch format {
	case "unix":
		return ok(time.Unix(t.Unix(), 0).String())
	default:
		return ok(t.Format(format))

}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text parameter is required")
	return ok(text)
}
}
}