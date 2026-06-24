package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "http://localhost:3000/tools"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}

func HandleCallTool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	toolName, _ :=getString(args, "toolName")
	if toolName == "" {
		return err("toolName is required")
}

	url := fmt.Sprintf("http://localhost:3000/tools/%s", toolName)
	bodyArgs, _ := json.Marshal(args)
	resp, e := http.DefaultClient.Post(url, "application/json", io.NopCloser(bytes.NewReader(bodyArgs)))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return success(string(data))
}