package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleSearchNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	resp, e := http.DefaultClient.Get("https://api.example.com/search?q=" + query)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	return success(string(body))
}

func HandleGetNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	resp, e := http.DefaultClient.Get("https://api.example.com/note/" + id)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var note map[string]interface{}
	if e := json.Unmarshal(body, &note); e != nil {
		return err(e.Error())
}

	return success(string(body))
}