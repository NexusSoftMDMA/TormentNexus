package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListBoards(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	token, _ :=getString(args, "token")
	if key == "" || token == "" {
		return err("API key and token required")
}

	url := fmt.Sprintf("https://api.trello.com/1/members/me/boards?key=%s&token=%s", key, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch boards: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Trello API error: %s", string(body)))
}

	var boards []map[string]interface{}
	if e := json.Unmarshal(body, &boards); e != nil {
		return err("failed to parse boards")
}

	return ok(fmt.Sprintf("Found %d boards", len(boards)))
}

func HandleListLists(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	token, _ :=getString(args, "token")
	boardID, _ :=getString(args, "boardId")
	if key == "" || token == "" || boardID == "" {
		return err("key, token, and boardId required")
}

	url := fmt.Sprintf("https://api.trello.com/1/boards/%s/lists?key=%s&token=%s", boardID, key, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch lists: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Trello API error: %s", string(body)))
}

	var lists []map[string]interface{}
	if e := json.Unmarshal(body, &lists); e != nil {
		return err("failed to parse lists")
}

	return ok(fmt.Sprintf("Found %d lists", len(lists)))
}