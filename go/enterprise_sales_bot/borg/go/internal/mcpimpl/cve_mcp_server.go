package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCveInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cveId, _ :=getString(args, "cve")
	if cveId == "" {
		return err("cve parameter is required")
}

	url := fmt.Sprintf("https://services.nvd.nist.gov/rest/json/cves/2.0?cveId=%s", cveId)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch CVE data: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP %d", resp.StatusCode))
}

	var data map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err("failed to parse response: " + e.Error())
}

	vulnerabilities, found := data["vulnerabilities"].([]interface{})
	if !found || len(vulnerabilities) == 0 {
		return err("CVE not found")
}

	vuln := vulnerabilities[0].(map[string]interface{})
	cve, found := vuln["cve"].(map[string]interface{})
	if !found {
		return err("unexpected response structure")
}

	descriptions, found := cve["descriptions"].([]interface{})
	if !found || len(descriptions) == 0 {
		return success("CVE found but no description available")
}

	desc := descriptions[0].(map[string]interface{})
	value, found := desc["value"].(string)
	if !found {
		return success("CVE found but description missing")
}

	return success(fmt.Sprintf("CVE %s: %s", cveId, value))
}