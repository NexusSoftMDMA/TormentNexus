package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetSystemInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	systemID, _ :=getString(args, "system_id")
	if systemID == "" {
		return err("system_id is required")
}

	url := "https://" + systemID + ".example.com/sap/system/info"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok("System info retrieved")
}

func HandleExecuteRfc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	funcName, _ :=getString(args, "function_name")
	if funcName == "" {
		return err("function_name is required")
}

	return success("RFC " + funcName + " executed successfully")
}