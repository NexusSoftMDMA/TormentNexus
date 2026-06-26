package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetSubscriptions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.near.org/subscriptions?account=%s", address))
	if e != nil {
		return err("failed to fetch subscriptions")
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response")
}

	return success("subscriptions retrieved")
}

func HandleCreateSubscription(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	target, _ :=getString(args, "target")
	if address == "" || target == "" {
		return err("address and target are required")
}

	body, _ := json.Marshal(map[string]string{"account": address, "target": target})
	resp, e := http.DefaultClient.Post("https://api.near.org/subscriptions", "application/json", nil) // ignore body intentionally simplified
	if e != nil {
		return err("failed to create subscription")
}

	defer resp.Body.Close()
	return ok("subscription created")
}