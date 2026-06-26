package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleQueryAccounts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "accessToken")
	url, _ :=getString(args, "environmentUrl")
	if token == "" || url == "" {
		return err("missing accessToken or environmentUrl")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url+"/api/data/v9.2/accounts", nil)
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
		return err("read failed: " + e.Error())
	}
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)))
	}
	var result interface{}
	json.Unmarshal(body, &result)
	return ok(fmt.Sprintf("query result: %v", result))
}

func HandleCreateRecord(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "accessToken")
	url, _ :=getString(args, "environmentUrl")
	entity, _ :=getString(args, "entityName")
	if token == "" || url == "" || entity == "" {
		return err("missing accessToken, environmentUrl, or entityName")
	}
	payload, e := json.Marshal(args["data"])
	if e != nil {
		return err("marshal failed: " + e.Error())
	}
	req, e := http.NewRequestWithContext(ctx, "POST", url+"/api/data/v9.2/"+entity, bytes.NewReader(payload))
	if e != nil {
		return err("request create failed: " + e.Error())
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	}
	if resp.StatusCode != 201 && resp.StatusCode != 204 {
		return err(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)))
	}
	return success("record created")
}