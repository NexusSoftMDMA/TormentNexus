package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchMemex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	url := fmt.Sprintf("http://localhost:3000/search?q=%s", q)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("search request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("unmarshal failed: %v", e))
}

	return ok(fmt.Sprintf("search results: %v", result))
}

func HandleGetMemexNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := fmt.Sprintf("http://localhost:3000/notes/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("note request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var note map[string]interface{}
	if e := json.Unmarshal(body, &note); e != nil {
		return err(fmt.Sprintf("unmarshal failed: %v", e))
}

	return ok(fmt.Sprintf("note: %v", note))
}