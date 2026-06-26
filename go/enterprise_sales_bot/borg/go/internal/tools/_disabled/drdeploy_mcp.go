package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleScanSite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	siteID, _ :=getString(args, "siteID")
	if siteID == "" {
		return err("siteID is required")
	}
	domain, _ :=getString(args, "domain")
	body := map[string]string{"siteID": siteID, "domain": domain}
	b, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal: " + e.Error())
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.drdeploy.com/scan", bytes.NewReader(b))
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
	}
	return success("scan started successfully")
}

func HandleGetFindings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	siteID, _ :=getString(args, "siteID")
	if siteID == "" {
		return err("siteID is required")
	}
	url := fmt.Sprintf("https://api.drdeploy.com/sites/%s/findings", siteID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
	}
	var findings interface{}
	e = json.Unmarshal(respBody, &findings)
	if e != nil {
		return err("failed to parse findings: " + e.Error())
	}
	return ok(string(respBody))
}