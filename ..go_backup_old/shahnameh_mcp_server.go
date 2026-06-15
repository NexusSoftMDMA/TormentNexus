package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetVerse(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.shahnameh.com/random")
	if e != nil {
		return err("failed to fetch verse: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	var result struct {
		Verse string `json:"verse"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(result.Verse)
}

func HandleGetVerseByPoet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	poet, _ :=getString(args, "poet")
	if poet == "" {
		return err("poet is required")
}

	url := "https://api.shahnameh.com/verses?poet=" + poet
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch verses: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	var result []string
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(result)
}