package tools

import (
	"context"
	"fmt"
)

var items = make(map[string]int)

func HandleGetTotal(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	total := 0
	for _, price := range items {
		total += price
	}
	return success("Total: " + fmt.Sprint(total))
}

func HandleAddItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	price, _ :=getInt(args, "price")
	items[name] = price
	return success("Added " + name + " with price " + fmt.Sprint(price))
}