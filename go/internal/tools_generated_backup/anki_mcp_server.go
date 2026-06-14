package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListDecks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	body := map[string]interface{}{
		"action":  "deckNames",
		"version": 6,
	}
	reqB, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.Post("http://localhost:8765", "application/json", asReadCloser(reqB))
	if e != nil {
		return err("Anki request failed: " + e.Error())
}

	defer resp.Body.Close()
	respB, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Result []string `json:"result"`
		Error  *string  `json:"error"`
	}
	if e := json.Unmarshal(respB, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if result.Error != nil {
		return err("Anki error: " + *result.Error)
}

	output := fmt.Sprintf("Decks: %v", result.Result)
	return ok(output)
}

func HandleCreateNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deckName, _ :=getString(args, "deckName")
	front, _ :=getString(args, "front")
	back, _ :=getString(args, "back")
	if deckName == "" || front == "" || back == "" {
		return err("deckName, front, and back are required")
}

	body := map[string]interface{}{
		"action":  "addNote",
		"version": 6,
		"params": map[string]interface{}{
			"note": map[string]interface{}{
				"deckName":  deckName,
				"modelName": "Basic",
				"fields": map[string]interface{}{
					"Front": front,
					"Back":  back,
				},
			},
		},
	}
	reqB, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.Post("http://localhost:8765", "application/json", asReadCloser(reqB))
	if e != nil {
		return err("Anki request failed: " + e.Error())
}

	defer resp.Body.Close()
	respB, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Result interface{} `json:"result"`
		Error  *string     `json:"error"`
	}
	if e := json.Unmarshal(respB, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if result.Error != nil {
		return err("Anki error: " + *result.Error)
}

	return ok("Note created successfully")
}

func asReadCloser(b []byte) io.ReadCloser {
	return &readCloser{Reader: asReader(b)}
}

type readCloser struct {
	io.Reader
}

func (rc *readCloser) Close() error { return nil }

func asReader(b []byte) io.Reader {
	return asByteReader(b)
}

type byteReader struct {
	data []byte
	pos  int
}

func (br *byteReader) Read(p []byte) (int, error) {
	if br.pos >= len(br.data) {
		return 0, io.EOF
	}
	n := copy(p, br.data[br.pos:])
	br.pos += n
	return n, nil
}