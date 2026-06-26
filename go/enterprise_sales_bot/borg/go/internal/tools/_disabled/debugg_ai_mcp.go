package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func HandleAsk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	resp, e := http.DefaultClient.Get("https://api.debugg.ai/ask?q=" + url.QueryEscape(query))
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("parse failed")
}

	answer, found := result["answer"].(string)
	if !found {
		return err("no answer")
}

	return ok(answer)
}