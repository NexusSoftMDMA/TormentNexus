package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleParseQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	resp, e := http.DefaultClient.Get("https://woogoro.com/api/hvac-estimate?action=parse&q=" + text)
	if e != nil {
		return err("failed to call API")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return ok(result)
}

func HandleCheckErrors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	quote, _ :=getString(args, "quote")
	resp, e := http.DefaultClient.Get("https://woogoro.com/api/hvac-estimate?action=check&q=" + quote)
	if e != nil {
		return err("failed to call API")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return ok(result)
}

func HandleLookupAveragePrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	equipment, _ :=getString(args, "equipment")
	resp, e := http.DefaultClient.Get("https://woogoro.com/api/hvac-estimate?action=price&q=" + equipment)
	if e != nil {
		return err("failed to call API")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return ok(result)
}

func HandleDraftDispute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issue, _ :=getString(args, "issue")
	resp, e := http.DefaultClient.Get("https://woogoro.com/api/hvac-estimate?action=dispute&q=" + issue)
	if e != nil {
		return err("failed to call API")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return ok(result)
}

func HandleNegotiationScript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	quote, _ :=getString(args, "quote")
	resp, e := http.DefaultClient.Get("https://woogoro.com/api/hvac-estimate?action=negotiate&q=" + quote)
	if e != nil {
		return err("failed to call API")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return ok(result)
}