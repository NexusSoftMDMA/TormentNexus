package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGenerateMeme(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("missing prompt")
}

	resp, e := http.DefaultClient.Get("https://api.memvid.example.com/generate?prompt=" + prompt)
	if e != nil {
		return err("API call failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var r struct{ URL string `json:"url"` }
	if e := json.Unmarshal(body, &r); e != nil {
		return err("parse failed: " + e.Error())
}

	return success(r.URL)
}

func HandleGetMemeStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing id")
}

	resp, e := http.DefaultClient.Get("https://api.memvid.example.com/status?id=" + id)
	if e != nil {
		return err("API call failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var r struct{ Status string `json:"status"` }
	if e := json.Unmarshal(body, &r); e != nil {
		return err("parse failed: " + e.Error())
}

	return success(r.Status)
}