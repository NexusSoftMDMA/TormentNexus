package mcpimpl

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleZen(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := http.DefaultClient
	resp, e := client.Get("https://api.github.com/zen")
	if e != nil {
		return err("failed to fetch zen: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	quote := strings.TrimSpace(string(body))
	return ok(quote)
}