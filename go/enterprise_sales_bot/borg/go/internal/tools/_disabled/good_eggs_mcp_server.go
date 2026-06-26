package tools

import (
	"context"
	"net/http"
)

func HandleSearchItems(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	_, e := http.DefaultClient.Get("https://api.goodeggs.com/v2/items?q=" + query)
	if e != nil {
		return err("search failed: " + e.Error())
}

	return ok("searched for: " + query)
}

func HandleAddToCart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	itemID, _ :=getString(args, "item_id")
	qty, _ :=getInt(args, "quantity")
	if qty < 1 {
		qty = 1
	}
	_, e := http.DefaultClient.Post("https://api.goodeggs.com/v2/cart/items", "application/json", nil)
	if e != nil {
		return err("add to cart failed: " + e.Error())
}

	return ok("added " + itemID + " x" + string(rune(qty)))
}