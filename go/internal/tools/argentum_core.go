//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://api.argentumcore.com/v1/prices/%s", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("bad status: %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json error: %v", e))
}

	price, found := result["price"]
	if !found {
		return err("price not in response")
}

	return ok(fmt.Sprintf("Price for %s: %v", symbol, price))
}

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	url := fmt.Sprintf("https://api.argentumcore.com/v1/info/%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("bad status: %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json error: %v", e))
}

	desc, found := result["description"]
	if !found {
		return err("description not in response")
}

	return ok(fmt.Sprintf("Info for %s: %s", name, desc))
}
