package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetKeywordSuggestions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword is required")
}

	url := fmt.Sprintf("https://api.storesignal.com/v1/keywords/suggest?api_key=%s&query=%s", os.Getenv("STORESIGNAL_KEY"), keyword)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch suggestions: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Keyword suggestions for '%s': %v", keyword, result))
}

func HandleGetAppDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appId, _ :=getString(args, "appId")
	if appId == "" {
		return err("appId is required")
}

	url := fmt.Sprintf("https://api.storesignal.com/v1/apps/details?api_key=%s&app_id=%s", os.Getenv("STORESIGNAL_KEY"), appId)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch app details: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Details for app '%s': %v", appId, result))
}