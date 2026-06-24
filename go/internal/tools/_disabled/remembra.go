package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleRemember(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	note, _ :=getString(args, "note")
	if note == "" {
		return err("note is required")
}

	body, _ := json.Marshal(map[string]string{"note": note})
	resp, e := http.DefaultClient.Post("http://localhost:8080/remembra", "application/json", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	return ok("remembered")
}

func HandleRecall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	resp, e := http.DefaultClient.Get("http://localhost:8080/remembra?q=" + query)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	notes, found := result["notes"]
	if !found {
		return success("no notes found")
}

	return success(fmt.Sprintf("found: %v", notes))
}