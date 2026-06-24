package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchBusiness(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	location, _ :=getString(args, "location")
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://api.openmart.com/search?q=%s&location=%s&limit=%d", url.QueryEscape(query), url.QueryEscape(location), limit)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("search failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse JSON failed: " + e.Error())
}

	return ok(fmt.Sprintf("Search results: %v", result))
}

func HandleEnrichDecisionMaker(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	businessName, _ :=getString(args, "businessName")
	domain, _ :=getString(args, "domain")
	u := fmt.Sprintf("https://api.openmart.com/enrich?business_name=%s&domain=%s", url.QueryEscape(businessName), url.QueryEscape(domain))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("enrichment failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse JSON failed: " + e.Error())
}

	return ok(fmt.Sprintf("Decision maker info: %v", result))
}