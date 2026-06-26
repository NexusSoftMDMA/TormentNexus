package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleOrderCoffee(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	coffee, _ :=getString(args, "type")
	size, _ :=getString(args, "size")
	body, _ := json.Marshal(map[string]string{"type": coffee, "size": size})
	resp, e := http.DefaultClient.Post("https://api.curless.com/order/coffee", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to order coffee: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("curless returned status " + resp.Status)
}

	return success("coffee ordered")
}

func HandleBookFlight(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")
	date, _ :=getString(args, "date")
	passengers, _ :=getInt(args, "passengers")
	payload, _ := json.Marshal(map[string]interface{}{"from": from, "to": to, "date": date, "passengers": passengers})
	resp, e := http.DefaultClient.Post("https://api.curless.com/book/flight", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("failed to book flight: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("curless returned status " + resp.Status)
}

	return success("flight booked")
}