package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tag, _ :=getString(args, "tag")
	url := "https://api.quotable.io/random"
	if tag != "" {
		url += "?tags=" + tag
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch quote: " + e.Error())
}

	defer resp.Body.Close()

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	content, found := data["content"].(string)
	if !found {
		return err("missing content in response")
}

	author, _ := data["author"].(string)
	quote := content
	if author != "" {
		quote += " - " + author
	}
	return success(quote)
}