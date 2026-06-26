package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleAnkiAction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	paramsStr, _ :=getString(args, "params")
	var params interface{}
	if paramsStr != "" {
		json.Unmarshal([]byte(paramsStr), &params)

	reqBody, e := json.Marshal(map[string]interface{}{
		"action":  action,
		"version": 6,
		"params":  params,
	})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post("http://127.0.0.1:8765", "application/json", bytes.NewReader(reqBody))
	if e != nil {
		return err("failed to call AnkiConnect: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if found, _ := result["error"]; found != nil && found.(string) != "" {
		return err("Anki error: " + found.(string))
}

	return ok("success")
}
}