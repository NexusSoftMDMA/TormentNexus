package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetFreightRate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	origin, _ :=getString(args, "origin")
	dest, _ :=getString(args, "destination")
	ct, _ :=getString(args, "container_type")
	url := "https://api.cerebrochain.io/v1/freight?origin=" + origin + "&dest=" + dest + "&container=" + ct
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch freight rate: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
	}
	rate, found := data["rate"].(float64)
	if !found {
		return err("rate not found in response")
	}
	return ok("Freight rate: " + fmt.Sprintf("%.2f", rate))
}

func HandleGetCommodityPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	url := "https://api.cerebrochain.io/v1/commodity?symbol=" + symbol
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch commodity price: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
	}
	price, found := data["price"].(float64)
	if !found {
		return err("price not found in response")
	}
	return ok("Price: " + fmt.Sprintf("%.2f", price))
}