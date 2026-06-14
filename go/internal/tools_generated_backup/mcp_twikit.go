package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetTweet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing tweet id")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.twitter.com/2/tweets/"+id, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return success("tweet retrieved")
}

func HandleSearchTweets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.twitter.com/2/tweets/search/recent?query="+query, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return success("search completed")
}// touch 1781132134
