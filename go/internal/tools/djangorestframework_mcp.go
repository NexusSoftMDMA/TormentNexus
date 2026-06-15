//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	}
	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("json parse error: " + e.Error())
	}
	return ok("fetched " + url)
}

func HandleCreate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
	}
	payload, _ :=getString(args, "payload")
	if payload == "" {
		return err("payload is required")
	}
	body := strings.NewReader(payload)
	resp, e := http.DefaultClient.Post(url, "application/json", body)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	}
	return success(string(respBody))
}
