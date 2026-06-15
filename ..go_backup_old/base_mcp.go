package tools

import (
	"context"
	"net/http"
	"strconv"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	return success(strconv.Itoa(sum))
}

func getString(args map[string]interface{}, key string) string {
	v, found := args[key]
	if !found {
		return ""
	}
	s, found := v.(string)
	if !found {
		return ""
	}
	return s
}

func getInt(args map[string]interface{}, key string) int {
	v, found := args[key]
	if !found {
		return 0
	}
	f, found := v.(float64)
	if !found {
		return 0
	}
	return int(f)
}