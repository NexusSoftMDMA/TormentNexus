package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleAddMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("missing content")
}

	data, e := json.Marshal(map[string]string{"content": content})
	if e != nil {
		return err("marshal failed")
}

	resp, e := http.DefaultClient.Post("http://localhost:8080/memory", "application/json", bytes.NewReader(data))
	if e != nil {
		return err("post failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("bad status: " + fmt.Sprint(resp.StatusCode))
}

	return ok("memory added")
}

func HandleSearchMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query")
}

	resp, e := http.DefaultClient.Get("http://localhost:8080/memory?q=" + query)
	if e != nil {
		return err("get failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	return ok(string(body))
}