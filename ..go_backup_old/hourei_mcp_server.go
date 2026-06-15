package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchLaws(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword is required")
}

	apiURL := fmt.Sprintf("https://elaws.e-gov.go.jp/api/1/lawlists?searchWord=%s", url.QueryEscape(keyword))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read: %v", e))
}

	var result interface{}
	json.Unmarshal(body, &result)
	return ok(fmt.Sprintf("Search results: %s", string(body)))
}

func HandleGetLaw(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lawId, _ :=getString(args, "lawId")
	if lawId == "" {
		return err("lawId is required")
}

	apiURL := fmt.Sprintf("https://elaws.e-gov.go.jp/api/1/lawdata?lawId=%s", url.QueryEscape(lawId))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read: %v", e))
}

	return ok(fmt.Sprintf("Law data: %s", string(body)))
}