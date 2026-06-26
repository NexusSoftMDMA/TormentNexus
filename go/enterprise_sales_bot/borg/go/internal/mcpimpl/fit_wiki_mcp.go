package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleGetFitWikiPage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		return err("topic is required")
}

	apiURL := "https://en.wikipedia.org/api/rest_v1/page/summary/" + url.PathEscape(topic)
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}// touch 1781132126
