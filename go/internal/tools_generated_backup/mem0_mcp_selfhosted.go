package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleAddMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	userID, _ :=getString(args, "user_id")
	if content == "" || userID == "" {
		return err("content and user_id required")
	}
	body, e := json.Marshal(map[string]string{"content": content, "user_id": userID})
	if e != nil {
		return err("failed to marshal")
	}
	resp, e := http.DefaultClient.Post("http://localhost:8000/memories", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: "+e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err("unexpected status: "+resp.Status)
	}
	return ok("memory added")
}

func HandleGetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "user_id")
	if userID == "" {
		return err("user_id required")
	}
	url := fmt.Sprintf("http://localhost:8000/memories/%s", userID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: "+e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: "+resp.Status)
	}
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed")
	}
	return success(fmt.Sprintf("%v", result))
}