package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleAddNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deck, _ :=getString(args, "deck")
	model, _ :=getString(args, "model")
	front, _ :=getString(args, "front")
	back, _ :=getString(args, "back")
	payload, e := json.Marshal(map[string]interface{}{
		"action":  "addNote",
		"version": 6,
		"params": map[string]interface{}{
			"note": map[string]interface{}{
				"deckName":  deck,
				"modelName": model,
				"fields":    map[string]string{"Front": front, "Back": back},
			},
		},
	})
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post("http://localhost:8765", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("failed to connect to Anki")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	if found, _ := result["error"]; found != nil {
		return err(found.(string))
}

	return ok("note added successfully")
}

func HandleGetDecks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	payload, e := json.Marshal(map[string]interface{}{
		"action":  "deckNames",
		"version": 6,
	})
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post("http://localhost:8765", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("failed to connect to Anki")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	if found, _ := result["error"]; found != nil {
		return err(found.(string))
}

	decks, found := result["result"].([]interface{})
	if !found {
		return err("unexpected response format")
}

	return success(decks)
}