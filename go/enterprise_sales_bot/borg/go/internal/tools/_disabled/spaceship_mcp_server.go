package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleListDNSRecords(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	apiKey, _ :=getString(args, "api_key")
	if domain == "" || apiKey == "" {
		return err("domain and api_key are required")
}

	recordType, _ :=getString(args, "type")
	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.spaceship.dev/dns/v1/%s/records", domain), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	if recordType != "" {
		q := req.URL.Query()
		q.Add("type", recordType)
		req.URL.RawQuery = q.Encode()

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	records, found := result["records"]
	if !found {
		return ok("no records found")
}

	return success(fmt.Sprintf("records: %v", records))
}

}

func HandleCreateDNSRecord(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	apiKey, _ :=getString(args, "api_key")
	name, _ :=getString(args, "name")
	recordType, _ :=getString(args, "type")
	value, _ :=getString(args, "value")
	ttl, _ :=getInt(args, "ttl")
	if domain == "" || apiKey == "" || name == "" || recordType == "" || value == "" {
		return err("domain, api_key, name, type, value are required")
}

	payload := map[string]interface{}{
		"name":  name,
		"type":  recordType,
		"value": value,
	}
	if ttl > 0 {
		payload["ttl"] = ttl
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://api.spaceship.dev/dns/v1/%s/records", domain), strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err("unexpected status: " + resp.Status)
}

	return success("record created")
}