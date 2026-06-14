package tools

import (
	"context"
	"net/http"
)

func HandleQueryMessages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	_ = query
	_ = limit
	resp, e := http.DefaultClient.Get("https://api.example.com/imessage?q=" + query)
	if e != nil {
		return err("failed to query iMessages: " + e.Error())
}

	defer resp.Body.Close()
	return success("found messages")
}

func HandleListConversations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = args
	resp, e := http.DefaultClient.Get("https://api.example.com/imessage/conversations")
	if e != nil {
		return err("failed to list conversations: " + e.Error())
}

	defer resp.Body.Close()
	return ok("conversations retrieved")
}