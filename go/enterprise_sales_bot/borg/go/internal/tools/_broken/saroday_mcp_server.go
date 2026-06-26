package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleCalculateSaju(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	body := map[string]string{
		"birthYear":  getString(args, "birthYear"),
		"birthMonth": getString(args, "birthMonth"),
		"birthDay":   getString(args, "birthDay"),
		"birthHour":  getString(args, "birthHour"),
		"gender":     getString(args, "gender"),
	}
	b, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://api.saroday.dev/saju", "application/json", strings.NewReader(string(b)))
	if e != nil {
		return err("api call failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(result)
}

func HandleLookupGlossary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	term, _ :=getString(args, "term")
	url := "https://api.saroday.dev/glossary?term=" + term
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("api call failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(result)
}