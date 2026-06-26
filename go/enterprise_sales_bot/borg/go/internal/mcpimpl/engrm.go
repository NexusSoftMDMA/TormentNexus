package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleCapture_engrm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	body, _ := json.Marshal(map[string]string{"key": key, "value": value})
	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(body))
	if e != nil {
		return err("capture failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("capture returned " + resp.Status)
}

	return ok("captured successfully")
}

func HandleContinuity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	sessionID, _ :=getString(args, "session_id")
	body, _ := json.Marshal(map[string]string{"session_id": sessionID})
	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(body))
	if e != nil {
		return err("continuity failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("continuity returned " + resp.Status)
}

	return success("continuity established")
}