package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchDrugs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	urlStr := "https://api.fda.gov/drug/drugsfda.json?search=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(urlStr)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Found %v results", result["meta"].(map[string]interface{})["results"].(map[string]interface{})["total"]))
}

func HandleGetDrugInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	drug, _ :=getString(args, "drug_name")
	if drug == "" {
		return err("drug_name is required")
}

	urlStr := "https://api.fda.gov/drug/label.json?search=openfda.brand_name:" + url.QueryEscape(drug)
	resp, e := http.DefaultClient.Get(urlStr)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Drug info retrieved: %v", result))
}