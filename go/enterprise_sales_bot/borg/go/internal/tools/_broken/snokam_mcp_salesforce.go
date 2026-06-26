package mcpimpl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleSalesforceQuery_snokam_mcp_salesforce(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	endpoint := os.Getenv("SF_ENDPOINT")
	token := os.Getenv("SF_ACCESS_TOKEN")
	if endpoint == "" || token == "" {
		return err("missing Salesforce credentials")
}

	u, e := url.Parse(fmt.Sprintf("%s/services/data/v58.0/query", endpoint))
	if e != nil {
		return err(e.Error())
}

	u.RawQuery = url.Values{"q": {q}}.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Salesforce error %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

func HandleSalesforceDescribe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	obj, _ :=getString(args, "object")
	if obj == "" {
		return err("object is required")
}

	endpoint := os.Getenv("SF_ENDPOINT")
	token := os.Getenv("SF_ACCESS_TOKEN")
	if endpoint == "" || token == "" {
		return err("missing Salesforce credentials")
}

	u := fmt.Sprintf("%s/services/data/v58.0/sobjects/%s/describe", endpoint, obj)
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Salesforce error %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}