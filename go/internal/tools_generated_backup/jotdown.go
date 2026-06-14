package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func HandleListNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:8080/notes")
	if e != nil {
		return err("failed to fetch notes: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleCreateNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	if title == "" || content == "" {
		return err("title and content are required")
}

	payload := map[string]string{"title": title, "content": content}
	jsonBytes, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	resp, e := http.DefaultClient.Post("http://localhost:8080/notes", "application/json", strings.NewReader(string(jsonBytes)))
	if e != nil {
		return err("failed to create note: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}