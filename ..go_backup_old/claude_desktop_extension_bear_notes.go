package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearchBearNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := fmt.Sprintf("http://localhost:8080/api/notes?query=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search notes: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	jsonBytes, _ := json.Marshal(result)
	return ok(string(jsonBytes))
}

func HandleCreateBearNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	body, _ :=getString(args, "body")
	payload, _ := json.Marshal(map[string]string{"title": title, "body": body})
	resp, e := http.DefaultClient.Post("http://localhost:8080/api/notes", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("failed to create note: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	jsonBytes, _ := json.Marshal(result)
	return success("note created: " + string(jsonBytes))
}