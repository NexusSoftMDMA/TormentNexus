package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleCreateOrder(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "baseUrl")
	if base == "" {
		return err("baseUrl is required")
}

	url := base + "/orders"
	body := map[string]interface{}{}
	for k, v := range args {
		if k != "baseUrl" {
			body[k] = v
		}
	}
	data, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(data))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("API returned status " + resp.Status)
}

	return ok("order created successfully")
}

func HandleGetOrder(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "baseUrl")
	if base == "" {
		return err("baseUrl is required")
}

	id, _ :=getString(args, "orderId")
	if id == "" {
		return err("orderId is required")
}

	url := base + "/orders/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return err("order not found")
}

	if resp.StatusCode >= 400 {
		return err("API returned status " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok("order retrieved successfully")
}