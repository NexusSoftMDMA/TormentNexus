package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleCreateNote_longhand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	reqBody, _ := json.Marshal(map[string]string{"text": text})
	resp, e := http.DefaultClient.Post("https://longhand.example.com/notes", "application/json", bytes.NewReader(reqBody))
	if e != nil {
		return err("failed to create note: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("note creation failed with status: " + resp.Status)
}

	return ok("note created")
}