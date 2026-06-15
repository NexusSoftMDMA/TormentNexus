package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetCustomer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	customerID, _ :=getString(args, "customer_id")
	apiKey, _ :=getString(args, "api_key")
	if customerID == "" || apiKey == "" {
		return err("customer_id and api_key are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.custify.com/v1/customers/%s", customerID), nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}

func HandleListCustomers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	page, _ :=getInt(args, "page")
	if page < 1 {
		page = 1
	}
	perPage, _ :=getInt(args, "per_page")
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	url := fmt.Sprintf("https://api.custify.com/v1/customers?page=%d&per_page=%d", page, perPage)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}