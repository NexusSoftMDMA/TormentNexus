package tools

import (
	"context"
)

func HandleGetBaipiaoList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	if category == "" {
		return ok("All baipiao items: free coffee, free ebook, free socks")
}

	return ok("Baipiao items in category " + category + ": free coffee, free ebook")
}

func HandleGetBaipiaoItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	name, _ :=getString(args, "name")
	if name != "" {
		return ok("Item " + name + " (ID " + string(rune(id)) + ") is a free gift")
}

	return ok("Baipiao item ID " + string(rune(id)) + " is a free gift")
}