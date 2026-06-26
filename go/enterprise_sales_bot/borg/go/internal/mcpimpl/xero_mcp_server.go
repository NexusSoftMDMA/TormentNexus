package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleX_xero_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	tenantID, _ :=getString(args, "tenant_id")
	if token == "" || tenantID == "" {
		return err("missing access_token or tenant_id")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.xero.com/api.xro/2.0/Invoices", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Xero-tenant-id", tenantID)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	invoices, found := result["Invoices"].([]interface{})
	if !found || len(invoices) == 0 {
		return success("No invoices found")
}

	return ok(fmt.Sprintf("Found %d invoices", len(invoices)))
}