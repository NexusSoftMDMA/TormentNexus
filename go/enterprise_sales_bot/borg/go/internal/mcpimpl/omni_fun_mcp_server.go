package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleJoke(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	typ, _ :=getString(args, "type")
	url := "https://official-joke-api.appspot.com/jokes/random"
	if typ != "" {
		url = "https://official-joke-api.appspot.com/jokes/" + typ + "/random"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch joke: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Setup  string `json:"setup"`
		Punchline string `json:"punchline"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode joke: " + e.Error())
}

	return success(result.Setup + " - " + result.Punchline)
}

func HandleFact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://uselessfacts.jsph.pl/random.json?language=en")
	if e != nil {
		return err("failed to fetch fact: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Text string `json:"text"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode fact: " + e.Error())
}

	return success(result.Text)
}