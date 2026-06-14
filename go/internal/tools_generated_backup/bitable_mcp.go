package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListSheets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	_ = name
	url := fmt.Sprintf("https://api.feishu.cn/open-apis/bitable/%s/tables", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to request tables")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	data, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result")
}

	return ok(string(data))
}

func HandleGetSheetData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sheetName, _ :=getString(args, "sheetName")
	if sheetName == "" {
		return err("sheetName is required")
}

	appToken, _ :=getString(args, "appToken")
	url := fmt.Sprintf("https://api.feishu.cn/open-apis/bitable/%s/tables/%s/records", appToken, sheetName)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to request records")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	data, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result")
}

	return ok(string(data))
}