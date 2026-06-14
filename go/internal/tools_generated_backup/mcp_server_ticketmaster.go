package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchEvents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	keyword, _ :=getString(args, "keyword")
	u, _ := url.Parse("https://app.ticketmaster.com/discovery/v2/events.json")
	q := u.Query()
	q.Set("apikey", apiKey)
	q.Set("keyword", keyword)
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleGetEventDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	eventID, _ :=getString(args, "eventId")
	u, _ := url.Parse(fmt.Sprintf("https://app.ticketmaster.com/discovery/v2/events/%s.json", eventID))
	q := u.Query()
	q.Set("apikey", apiKey)
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}