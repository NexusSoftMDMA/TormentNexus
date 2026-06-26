package mcpimpl

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleSpectrawlQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := "https://api.spectrawl.example/search?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to query: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}

func HandleSpectrawlPing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}