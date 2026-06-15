package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListBoards(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "apiKey")
	token, _ :=getString(args, "token")
	if key == "" || token == "" {
		return err("apiKey and token required")
}

	url := fmt.Sprintf("https://api.trello.com/1/members/me/boards?key=%s&token=%s", key, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d", resp.StatusCode))
}

	var boards []interface{}
	if e := json.NewDecoder(resp.Body).Decode(&boards); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success("Found boards")
}

func HandleGetBoard(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "apiKey")
	token, _ :=getString(args, "token")
	boardID, _ :=getString(args, "boardId")
	if key == "" || token == "" || boardID == "" {
		return err("apiKey, token, and boardId required")
}

	url := fmt.Sprintf("https://api.trello.com/1/boards/%s?key=%s&token=%s", boardID, key, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d", resp.StatusCode))
}

	var board map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&board); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success("Got board")
}