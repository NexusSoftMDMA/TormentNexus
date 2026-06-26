package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

func HandleSearchRecords(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	endpoint := "https://api.compoid.ai/search?q=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(endpoint)
	if e != nil {
		return err("failed to search records: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	records, found := result["records"]
	if !found {
		return err("no records found")
}

	data, _ := json.Marshal(records)
	return ok("Found records: " + string(data))
}

func HandleCreateEntry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	description, _ :=getString(args, "description")
	if name == "" {
		return err("name parameter is required")
}

	body := map[string]string{"name": name, "description": description}
	jsonBody, _ := json.Marshal(body)
	resp, e := http.DefaultClient.Post("https://api.compoid.ai/entries", "application/json", strings.NewReader(string(jsonBody)))
	if e != nil {
		return err("failed to create entry: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok("Entry created: " + result["id"].(string))
}