package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetWishlist(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "userId")
	if userID == "" {
		return err("userId is required")
}

	url := fmt.Sprintf("https://api.forme.gifts/wishlists/%s", userID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch wishlist: %v", e))
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid response")
}

	return ok(fmt.Sprintf("wishlist: %s", string(body)))
}

func HandleAddItem_formegifts_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "userId")
	itemName, _ :=getString(args, "itemName")
	if userID == "" || itemName == "" {
		return err("userId and itemName are required")
}

	payload := map[string]string{"itemName": itemName}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.forme.gifts/wishlists/%s/items", userID)
	resp, e := http.DefaultClient.Post(url, "application/json", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to add item: %v", e))
}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return success(fmt.Sprintf("added item: %s", string(respBody)))
}