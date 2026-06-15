package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchCVE(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword is required")
}

	searchURL := fmt.Sprintf("https://cve.circl.lu/api/cve/%s", url.QueryEscape(keyword))
	resp, e := http.DefaultClient.Get(searchURL)
	if e != nil {
		return err("failed to query API: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return success("Search results: " + string(data))
}

func HandleGetCVE(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cveID, _ :=getString(args, "cveId")
	if cveID == "" {
		return err("cveId is required")
}

	detailsURL := fmt.Sprintf("https://cve.circl.lu/api/cve/%s", cveID)
	resp, e := http.DefaultClient.Get(detailsURL)
	if e != nil {
		return err("failed to query API: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return success("CVE details: " + string(data))
}