package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleResourceGroups(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subscriptionId, _ :=getString(args, "subscriptionId")
	token, _ :=getString(args, "accessToken")
	if subscriptionId == "" || token == "" {
		return err("subscriptionId and accessToken are required")
}

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourcegroups?api-version=2021-04-01", subscriptionId)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return success(string(body))
}

func HandleVirtualMachines(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subscriptionId, _ :=getString(args, "subscriptionId")
	resourceGroup, _ :=getString(args, "resourceGroup")
	token, _ :=getString(args, "accessToken")
	if subscriptionId == "" || token == "" || resourceGroup == "" {
		return err("subscriptionId, resourceGroup, and accessToken are required")
}

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines?api-version=2022-03-01", subscriptionId, resourceGroup)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return success(string(body))
}