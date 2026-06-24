package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type jokeResponse struct {
	Value string `json:"value"`
}

func HandleRandomJoke(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.chucknorris.io/jokes/random")
	if e != nil {
		return err("failed to fetch joke: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	var joke jokeResponse
	if e := json.Unmarshal(body, &joke); e != nil {
		return err("failed to parse joke: " + e.Error())
}

	return ok(joke.Value)
}