package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetMeme(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	url := "https://meme-api.herokuapp.com/gimme"
	if text != "" {
		url = "https://api.memegen.link/images/custom/" + text + ".jpg"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch meme: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		URL string `json:"url"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(result.URL)
}