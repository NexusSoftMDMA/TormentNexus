package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchCompanies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	url := fmt.Sprintf("https://api.company-information.service.gov.uk/search/companies?q=%s", query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(apiKey, "")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jsonBytes))
}

func HandleGetCompany(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	companyNumber, _ :=getString(args, "company_number")
	if companyNumber == "" {
		return err("company_number is required")
}

	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	url := fmt.Sprintf("https://api.company-information.service.gov.uk/company/%s", companyNumber)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(apiKey, "")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jsonBytes))
}