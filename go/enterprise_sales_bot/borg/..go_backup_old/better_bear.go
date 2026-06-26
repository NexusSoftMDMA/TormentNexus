package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleCreateNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	body, _ :=getString(args, "body")
	if title == "" {
		return err("title is required")
}

	payload := map[string]string{"title": title, "body": body}
	data, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post("http://localhost:8080/notes", "application/json", bytes.NewReader(data))
	if e != nil {
		return err("failed to create note: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err("unexpected status: " + resp.Status)
}

	return success("note created")
}

func HandleSearchNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := "http://localhost:8080/notes?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	var notes []map[string]string
	if e := json.NewDecoder(resp.Body).Decode(&notes); e != nil {
		return err("failed to decode response")
}

	text, _ := json.Marshal(notes)
	return ok(string(text))
}