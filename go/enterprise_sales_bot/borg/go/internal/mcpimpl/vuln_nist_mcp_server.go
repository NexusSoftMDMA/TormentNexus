package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCve(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cveID, _ :=getString(args, "cveId")
	if cveID == "" {
		return err("missing cveId")
}

	url := fmt.Sprintf("https://services.nvd.nist.gov/rest/json/cves/2.0?cveId=%s", cveID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Vulnerabilities []struct {
			CVE struct {
				ID       string `json:"id"`
				Descriptions []struct {
					Lang  string `json:"lang"`
					Value string `json:"value"`
				} `json:"descriptions"`
				Metrics map[string]interface{} `json:"metrics"`
			} `json:"cve"`
		} `json:"vulnerabilities"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if len(result.Vulnerabilities) == 0 {
		return err("CVE not found")
}

	cve := result.Vulnerabilities[0].CVE
	desc := ""
	for _, d := range cve.Descriptions {
		if d.Lang == "en" {
			desc = d.Value
			break
		}
	}
	return ok(fmt.Sprintf("CVE: %s\nDescription: %s\nMetrics: %v", cve.ID, desc, cve.Metrics))
}

func HandleSearch_vuln_nist_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "keyword")
	if query == "" {
		return err("missing keyword")
}

	url := fmt.Sprintf("https://services.nvd.nist.gov/rest/json/cves/2.0?keywordSearch=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Vulnerabilities []struct {
			CVE struct {
				ID string `json:"id"`
			} `json:"cve"`
		} `json:"vulnerabilities"`
		TotalResults int `json:"totalResults"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	ids := ""
	for i, v := range result.Vulnerabilities {
		if i > 0 {
			ids += ", "
		}
		ids += v.CVE.ID
	}
	return ok(fmt.Sprintf("Found %d results: %s", result.TotalResults, ids))
}