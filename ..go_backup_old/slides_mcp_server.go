package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListDecks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = os.Getenv("SLIDES_API_HOST")
		if host == "" {
			host = "http://localhost:8080"
		}
	}
	resp, e := http.DefaultClient.Get(host + "/decks")
	if e != nil {
		return err("failed to list decks: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var decks []map[string]interface{}
	if e := json.Unmarshal(body, &decks); e != nil {
		return err("invalid JSON: " + e.Error())
}

	out, _ := json.Marshal(decks)
	return ok("decks: " + string(out))
}

func HandleCreateDeck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = os.Getenv("SLIDES_API_HOST")
		if host == "" {
			host = "http://localhost:8080"
		}
	}
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title is required")
}

	body := map[string]string{"title": title}
	jsonBody, _ := json.Marshal(body)
	resp, e := http.DefaultClient.Post(host+"/decks", "application/json", bytes.NewReader(jsonBody))
	if e != nil {
		return err("failed to create deck: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(fmt.Sprintf("created deck: %s", string(respBody)))
}