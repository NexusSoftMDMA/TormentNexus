package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiKey, _ :=getString(args, "apiKey")
	url := fmt.Sprintf("https://api.tripadvisor.com/api/v1/search?key=%s&query=%s", apiKey, query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("search results: %v", result))
}

func HandleGetReviews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	locationId, _ :=getString(args, "locationId")
	if locationId == "" {
		return err("locationId is required")
}

	apiKey, _ :=getString(args, "apiKey")
	url := fmt.Sprintf("https://api.tripadvisor.com/api/v1/reviews/%s?key=%s", locationId, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("reviews: %v", result))
}