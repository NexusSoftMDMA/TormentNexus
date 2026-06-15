package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// HandleSearchDatasets searches datasets on data.gouv.fr
func HandleSearchDatasets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiURL := fmt.Sprintf("https://www.data.gouv.fr/api/1/datasets/?q=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse response failed: " + e.Error())
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}

// HandleGetDataset retrieves a specific dataset by ID
func HandleGetDataset(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	apiURL := fmt.Sprintf("https://www.data.gouv.fr/api/1/datasets/%s/", url.PathEscape(id))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse response failed: " + e.Error())
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}