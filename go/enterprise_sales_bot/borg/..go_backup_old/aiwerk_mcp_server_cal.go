package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func HandleListEventTypes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
	}
	reqURL := fmt.Sprintf("https://api.cal.com/v1/event-types?apiKey=%s", url.QueryEscape(apiKey))
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
	}
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
	}
	return ok(string(body))
}

func HandleCreateBooking(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
	}
	eventTypeID, _ :=getInt(args, "eventTypeId")
	if eventTypeID == 0 {
		return err("eventTypeId is required")
	}
	start, _ :=getString(args, "start")
	end, _ :=getString(args, "end")
	name, _ :=getString(args, "name")
	email, _ :=getString(args, "email")
	if start == "" || end == "" || name == "" || email == "" {
		return err("start, end, name, email are required")
	}
	payload := fmt.Sprintf(`{"eventTypeId":%d,"start":"%s","end":"%s","responses":{"name":"%s","email":"%s"}}`,
		eventTypeID, start, end, name, email)
	reqURL := fmt.Sprintf("https://api.cal.com/v1/bookings?apiKey=%s", url.QueryEscape(apiKey))
	resp, e := http.DefaultClient.Post(reqURL, "application/json", strings.NewReader(payload))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
	}
	return success(strconv.Itoa(resp.StatusCode) + ": " + string(body))
}