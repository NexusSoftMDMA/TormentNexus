package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleCheckImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	serverUrl, _ :=getString(args, "serverUrl")
	imageBase64, _ :=getString(args, "imageBase64")
	appName, _ :=getString(args, "appName")
	testName, _ :=getString(args, "testName")

	body, _ := json.Marshal(map[string]string{
		"apiKey":      apiKey,
		"imageBase64": imageBase64,
		"appName":     appName,
		"testName":    testName,
	})
	req, e := http.NewRequestWithContext(ctx, "POST", serverUrl+"/api/images/check", bytes.NewBuffer(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok("check submitted: " + result["id"].(string))
}

func HandleGetResults(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	serverUrl, _ :=getString(args, "serverUrl")
	testId, _ :=getString(args, "testId")

	req, e := http.NewRequestWithContext(ctx, "GET", serverUrl+"/api/sessions/"+testId, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Api-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success("result: " + result["status"].(string))
}