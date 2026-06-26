package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGoogleAdsCampaigns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	customerID, _ :=getString(args, "customer_id")
	if customerID == "" {
		return err("customer_id is required")
	}
	url := fmt.Sprintf("https://googleads.googleapis.com/v18/customers/%s/campaigns", customerID)
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
	return success(string(body))
}

func HandleGA4Report(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	propertyID, _ :=getString(args, "property_id")
	if propertyID == "" {
		return err("property_id is required")
	}
	url := fmt.Sprintf("https://analyticsdata.googleapis.com/v1beta/properties/%s:runReport", propertyID)
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
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
	return success(string(body))
}