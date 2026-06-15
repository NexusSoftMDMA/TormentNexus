package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pokemon, _ :=getString(args, "pokemon")
	if pokemon == "" {
		return err("pokemon name is required")
}

	url := "https://pokeapi.co/api/v2/pokemon/" + pokemon
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch pokemon data")
}

	defer resp.Body.Close()

	var data map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err("failed to decode pokemon data")
}

	return success("Fetched data for " + pokemon)
}

func HandleY(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return success("Message received: " + message)
}