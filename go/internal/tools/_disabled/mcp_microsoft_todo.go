package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListTodos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	listID, _ :=getString(args, "list_id")
	token, _ :=getString(args, "token")
	if listID == "" || token == "" {
		return err("missing list_id or token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://graph.microsoft.com/v1.0/me/todo/lists/%s/tasks", listID), nil)
	if e != nil {
		return err("request create failed")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("unmarshal failed")
}

	raw, found := result["value"]
	if !found {
		return err("no tasks found")
}

	data, _ := json.Marshal(raw)
	return ok(string(data))
}