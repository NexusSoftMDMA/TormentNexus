//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func HandleGetCell(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sheet, _ :=getString(args, "sheet")
	if sheet == "" {
		return err("sheet is required")
}

	cell, _ :=getString(args, "cell")
	if cell == "" {
		return err("cell is required")
}

	u := fmt.Sprintf("https://api.example.com/spreadsheet/%s/%s", url.PathEscape(sheet), url.PathEscape(cell))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
}

	value, found := data["value"]
	if !found {
		return err("value not found")
	return ok(fmt.Sprintf("Cell %s value: %v", cell, value))
}

func HandleSetCell(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sheet, _ :=getString(args, "sheet")
	if sheet == "" {
		return err("sheet is required")
}

	cell, _ :=getString(args, "cell")
	if cell == "" {
		return err("cell is required")
}

	value, _ :=getString(args, "value")
	if value == "" {
		return err("value is required")
}

	body := fmt.Sprintf(`{"value":"%s"}`, strings.ReplaceAll(value, `"`, `\"`))
	u := fmt.Sprintf("https://api.example.com/spreadsheet/%s/%s", url.PathEscape(sheet), url.PathEscape(cell))
	resp, e := http.DefaultClient.Post(u, "application/json", strings.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("set failed with status " + resp.Status)
	return success("Cell updated successfully")
}
}
}
}