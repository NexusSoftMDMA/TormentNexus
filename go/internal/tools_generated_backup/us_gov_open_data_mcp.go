package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetBill(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	congress, _ :=getString(args, "congress")
	billType, _ :=getString(args, "billType")
	billNumber, _ :=getString(args, "billNumber")
	if congress == "" || billType == "" || billNumber == "" {
		return err("missing required parameters: congress, billType, billNumber")
}

	url := fmt.Sprintf("https://api.congress.gov/v3/bill/%s/%s/%s", congress, billType, billNumber)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch bill: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Bill data: %v", result))
}

func HandleSearchFDA(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
}

	url := fmt.Sprintf("https://api.fda.gov/drug/drugsfda.json?search=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to query FDA: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("FDA results: %v", result))
}