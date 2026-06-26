package mcpimpl

import (
	"context"
	"fmt"
	"net/http"
)

func HandleReadCoils(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	unitID, _ :=getInt(args, "unit_id")
	address, _ :=getInt(args, "address")
	count, _ :=getInt(args, "count")
	if baseURL == "" {
		return err("base_url is required")
	}
	url := fmt.Sprintf("%s/coils?unit=%d&address=%d&count=%d", baseURL, unitID, address, count)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("HTTP error: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("HTTP status: " + resp.Status)
	}
	return ok("Coils read successfully")
}

func HandleWriteSingleCoil(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	unitID, _ :=getInt(args, "unit_id")
	address, _ :=getInt(args, "address")
	value, _ :=getBool(args, "value")
	if baseURL == "" {
		return err("base_url is required")
	}
	val := 0
	if value {
		val = 1
	}
	url := fmt.Sprintf("%s/coil?unit=%d&address=%d&value=%d", baseURL, unitID, address, val)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("HTTP error: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("HTTP status: " + resp.Status)
	}
	return ok("Coil written successfully")
}