package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListRecords(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseId, _ :=getString(args, "baseId")
	if baseId == "" {
		return err("baseId required")
}

	tableId, _ :=getString(args, "tableId")
	if tableId == "" {
		return err("tableId required")
}

	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey required")
}

	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", baseId, tableId)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	records, found := result["records"]
	if !found {
		return err("no records field in response")
}

	return success(fmt.Sprintf("Found %d records", len(records.([]interface{}))))
}

func HandleCreateRecord(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseId, _ :=getString(args, "baseId")
	if baseId == "" {
		return err("baseId required")
}

	tableId, _ :=getString(args, "tableId")
	if tableId == "" {
		return err("tableId required")
}

	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey required")
}

	fields, found := args["fields"].(map[string]interface{})
	if !found {
		return err("fields must be an object")
}

	payload := map[string]interface{}{"fields": fields}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal fields: " + e.Error())
}

	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", baseId, tableId)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, io.NopCloser(bytes.NewReader(bodyBytes)))
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
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return err(fmt.Sprintf("API error: %s", string(respBody)))
}

	return success("Record created successfully")
}