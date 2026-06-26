package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleWorldsenseGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := "https://api.worldsense.example.com/data?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	return success("Worldsense data: " + string(body))
}