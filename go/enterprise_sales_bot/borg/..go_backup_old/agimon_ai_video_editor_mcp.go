package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleRenderComposition(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	compID, _ :=getString(args, "compositionId")
	if compID == "" {
		return err("compositionId is required")
	}
	inputProps, _ :=getString(args, "inputProps")
	url := "http://localhost:3000/render?compositionId=" + compID + "&inputProps=" + inputProps
	resp, e := http.DefaultClient.Post(url, "application/json", nil)
	if e != nil {
		return err("render request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse response failed: " + e.Error())
	}
	return success("render started")
}