package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

func HandleCreateMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	word, _ :=getString(args, "word")
	meaning, _ :=getString(args, "meaning")
	if word == "" || meaning == "" {
		return err("word and meaning required")
}

	body, _ := json.Marshal(map[string]string{"word": word, "meaning": meaning})
	resp, e := http.DefaultClient.Post("https://memory-plus.example.com/memories", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to create memory: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err("unexpected status: " + resp.Status)
}

	return ok("memory created")
}

func HandleRecallMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	word, _ :=getString(args, "word")
	if word == "" {
		return err("word required")
}

	resp, e := http.DefaultClient.Get("https://memory-plus.example.com/memories/" + word)
	if e != nil {
		return err("failed to recall: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("memory not found")
}

	var data map[string]string
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("invalid response")
}

	return success("Memory: " + data["meaning"])
}