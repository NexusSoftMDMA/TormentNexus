package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleGetSeries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key required")
}

	seriesID, _ :=getString(args, "series_id")
	if seriesID == "" {
		return err("series_id required")
}

	url := fmt.Sprintf("https://api.stlouisfed.org/fred/series/observations?series_id=%s&api_key=%s&file_type=json", seriesID, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("json error: %v", e))
}

	return success(fmt.Sprintf("%v", data))
}

func HandleSearchSeries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key required")
}

	searchText, _ :=getString(args, "search_text")
	if searchText == "" {
		return err("search_text required")
}

	url := fmt.Sprintf("https://api.stlouisfed.org/fred/series/search?search_text=%s&api_key=%s&file_type=json", searchText, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("json error: %v", e))
}

	return success(fmt.Sprintf("%v", data))
}// touch 1781132126
