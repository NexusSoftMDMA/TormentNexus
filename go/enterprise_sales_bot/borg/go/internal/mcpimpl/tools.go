package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleEcho_tools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return ok(fmt.Sprintf("Echo: %s", msg))
}

func HandleRandomJoke_tools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://official-joke-api.appspot.com/random_joke")
	if e != nil {
		return err("failed to fetch joke")
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse joke")
}

	setup, _ := data["setup"].(string)
	punchline, _ := data["punchline"].(string)
	joke := fmt.Sprintf("%s - %s", setup, punchline)
	return ok(joke)
}