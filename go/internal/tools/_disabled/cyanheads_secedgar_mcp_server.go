package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleGetFiling(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("missing ticker")
	}
	base := "https://data.sec.gov/submissions/CIK%s.json"
	cik, _ :=getString(args, "cik")
	if cik == "" {
		cik = "0000320193"
	}
	reqURL := fmt.Sprintf(base, cik)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if e != nil {
		return err(e.Error())
	}
	req.Header.Set("User-Agent", "MCP-Client/1.0")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
	}
	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
	}
	return ok(fmt.Sprintf("Retrieved data for %s", ticker))
}

func HandleSearchXBRL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query")
	}
	base := "https://www.sec.gov/cgi-bin/browse-edgar"
	params := url.Values{}
	params.Set("action", "getcompany")
	params.Set("CIK", query)
	params.Set("type", "")
	params.Set("count", "10")
	reqURL := base + "?" + params.Encode()
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if e != nil {
		return err(e.Error())
	}
	req.Header.Set("User-Agent", "MCP-Client/1.0")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
	}
	return ok(fmt.Sprintf("Search results for %s: %d bytes", query, len(body)))
}// touch 1781132124
