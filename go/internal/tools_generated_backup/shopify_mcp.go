package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListProducts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	shop, _ :=getString(args, "shop")
	token, _ :=getString(args, "access_token")
	url := fmt.Sprintf("https://%s.myshopify.com/admin/api/2023-01/products.json", shop)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request"), e
	}
	req.Header.Set("X-Shopify-Access-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request"), e
	}
	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response"), e
	}
	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON"), e
	}
	return success(fmt.Sprintf("%v", result))
}

func HandleGetOrder(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	shop, _ :=getString(args, "shop")
	token, _ :=getString(args, "access_token")
	orderID, _ :=getString(args, "order_id")
	url := fmt.Sprintf("https://%s.myshopify.com/admin/api/2023-01/orders/%s.json", shop, orderID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request"), e
	}
	req.Header.Set("X-Shopify-Access-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request"), e
	}
	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response"), e
	}
	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON"), e
	}
	return success(fmt.Sprintf("%v", result))
}