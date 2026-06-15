package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetSheet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sheetID, _ :=getString(args, "spreadsheetId")
	sheetRange, _ :=getString(args, "range")

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s", sheetID, sheetRange)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success(fmt.Sprintf("%v", result))
}