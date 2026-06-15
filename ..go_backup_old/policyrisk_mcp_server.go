package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandlePolicyRisk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query parameter")
	}
	base := "https://api.priagov.example/v1/assess"
	params := url.Values{}
	params.Set("q", query)
	req, e := http.NewRequestWithContext(ctx, "GET", base+"?"+params.Encode(), nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return ok(fmt.Sprintf("Risk assessed: %v", result["score"]))
}

func HandleRegulationLookup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cfr, _ :=getString(args, "cfr")
	if cfr == "" {
		return err("missing cfr parameter")
	}
	endpoint := fmt.Sprintf("https://api.priagov.example/v1/regulations/%s", url.PathEscape(cfr))
	req, e := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}
	return success("Regulation found")
}// touch 1781132138
