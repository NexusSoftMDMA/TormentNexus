package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListProducts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain := os.Getenv("SHOPIFY_STORE_DOMAIN")
	token := os.Getenv("SHOPIFY_STOREFRONT_ACCESS_TOKEN")
	query := `{ products(first: 10) { edges { node { id title handle } } } }`
	payload := map[string]string{"query": query}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://%s/api/2023-10/graphql.json", domain), bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Storefront-Access-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: "+e.Error())
	}
	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	if resp.StatusCode != 200 {
		return err("API error: "+string(respBody))
	}
	return success(string(respBody))
}

func HandleGetProduct(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	handle, _ :=getString(args, "handle")
	if handle == "" {
		return err("handle is required")
	}
	domain := os.Getenv("SHOPIFY_STORE_DOMAIN")
	token := os.Getenv("SHOPIFY_STOREFRONT_ACCESS_TOKEN")
	query := fmt.Sprintf(`{ productByHandle(handle: "%s") { id title description } }`, handle)
	payload := map[string]string{"query": query}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://%s/api/2023-10/graphql.json", domain), bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Storefront-Access-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: "+e.Error())
	}
	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	if resp.StatusCode != 200 {
		return err("API error: "+string(respBody))
	}
	return success(string(respBody))
}