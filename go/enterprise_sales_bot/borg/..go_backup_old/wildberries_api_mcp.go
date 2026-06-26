package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetOrders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	dateFrom, _ :=getString(args, "dateFrom")
	dateTo, _ :=getString(args, "dateTo")
	if apiKey == "" {
		return err("apiKey is required")
}

	url := fmt.Sprintf("https://statistics-api.wildberries.ru/api/v1/supplier/orders?key=%s&dateFrom=%s&dateTo=%s", apiKey, dateFrom, dateTo)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	return ok(string(body))
}

func HandleGetSales(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	dateFrom, _ :=getString(args, "dateFrom")
	dateTo, _ :=getString(args, "dateTo")
	if apiKey == "" {
		return err("apiKey is required")
}

	url := fmt.Sprintf("https://statistics-api.wildberries.ru/api/v1/supplier/sales?key=%s&dateFrom=%s&dateTo=%s", apiKey, dateFrom, dateTo)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	return ok(string(body))
}