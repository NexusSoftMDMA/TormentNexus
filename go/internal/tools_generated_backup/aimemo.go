package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

type Memo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func HandleListMemos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.aimemo.example/memos")
	if e != nil {
		return err("failed to fetch memos")
}

	defer resp.Body.Close()
	var memos []Memo
	if e := json.NewDecoder(resp.Body).Decode(&memos); e != nil {
		return err("failed to decode response")
}

	result, _ := json.Marshal(memos)
	return success(string(result))
}

func HandleGetMemo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := "https://api.aimemo.example/memos/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch memo")
}

	defer resp.Body.Close()
	var memo Memo
	if e := json.NewDecoder(resp.Body).Decode(&memo); e != nil {
		return err("failed to decode response")
}

	result, _ := json.Marshal(memo)
	return success(string(result))
}