package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleReadSheet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	spreadsheetId, _ :=getString(args, "spreadsheetId")
	sheetRange, _ :=getString(args, "range")
	apiKey, _ :=getString(args, "apiKey")
	if spreadsheetId == "" || sheetRange == "" || apiKey == "" {
		return err("missing required args: spreadsheetId, range, apiKey")
}

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s?key=%s", spreadsheetId, sheetRange, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse JSON failed: %v", e))
}

	values, found := data["values"]
	if !found {
		return err("no values in response")
}

	result, _ := json.Marshal(values)
	return ok(string(result))
}