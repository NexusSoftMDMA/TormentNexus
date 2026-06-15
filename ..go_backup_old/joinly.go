package tools

import (
	"context"
	"strings"
)

func HandleJoin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sep, _ :=getString(args, "separator")
	if sep == "" {
		sep = ","
	}
	rawItems, found := args["items"]
	if !found {
		return err("missing 'items' argument")
}

	items, found := rawItems.([]interface{})
	if !found {
		return err("'items' must be an array")
}

	strs := make([]string, len(items))
	for i, v := range items {
		s, found := v.(string)
		if !found {
			return err("each item must be a string")
}

		strs[i] = s
	}
	result := strings.Join(strs, sep)
	return ok(result)
}

func HandleReverseJoin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rawItems, found := args["items"]
	if !found {
		return err("missing 'items' argument")
}

	items, found := rawItems.([]interface{})
	if !found {
		return err("'items' must be an array")
}

	strs := make([]string, len(items))
	for i, v := range items {
		s, found := v.(string)
		if !found {
			return err("each item must be a string")
}

		runes := []rune(s)
		n := len(runes)
		for j := 0; j < n/2; j++ {
			runes[j], runes[n-1-j] = runes[n-1-j], runes[j]
		}
		strs[i] = string(runes)

	result := strings.Join(strs, getString(args, "separator"))
	return ok(result)
}
}