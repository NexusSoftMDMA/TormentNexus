package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleReadSheet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sheetID, _ :=getString(args, "spreadsheetId")
	sheetRange, _ :=getString(args, "range")
	token, _ :=getString(args, "accessToken")

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s", sheetID, sheetRange)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("API error: " + string(body))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(result)
}

func HandleWriteSheet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sheetID, _ :=getString(args, "spreadsheetId")
	sheetRange, _ :=getString(args, "range")
	token, _ :=getString(args, "accessToken")
	valuesJSON, _ :=getString(args, "values")

	var values [][]interface{}
	if e := json.Unmarshal([]byte(valuesJSON), &values); e != nil {
		return err("invalid values JSON: " + e.Error())
}

	payload := map[string]interface{}{"values": values}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s?valueInputOption=USER_ENTERED", sheetID, sheetRange)
	req, e := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(string(bodyBytes)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("API error: " + string(respBody))
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(result)
}