package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func TrackEvent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	event, _ :=getString(args, "event")
	properties := args["properties"]
	url, _ :=getString(args, "url")
	if url == "" {
		url = "https://api.agnost.ai/track"
	}
	body := map[string]interface{}{
		"event":      event,
		"properties": properties,
	}
	var buf bytes.Buffer
	if e := json.NewEncoder(&buf).Encode(body); e != nil {
		return err("failed to encode body")
}

	resp, e := http.DefaultClient.Post(url, "application/json", &buf)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	return ok("event tracked")
}

func IdentifyUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userId, _ :=getString(args, "userId")
	traits := args["traits"]
	url, _ :=getString(args, "url")
	if url == "" {
		url = "https://api.agnost.ai/identify"
	}
	body := map[string]interface{}{
		"userId": userId,
		"traits": traits,
	}
	var buf bytes.Buffer
	if e := json.NewEncoder(&buf).Encode(body); e != nil {
		return err("failed to encode body")
}

	resp, e := http.DefaultClient.Post(url, "application/json", &buf)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	return ok("user identified")
}