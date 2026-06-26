package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetPtEdgeInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing id parameter")
	}
	resp, e := http.DefaultClient.Get("https://api.ptedge.example.com/info?id=" + id)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
	}
	return ok("retrieved Pt Edge info: " + string(body))
}