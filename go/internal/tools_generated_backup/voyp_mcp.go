package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	text, _ :=getString(args, "text")
	if to == "" || text == "" {
		return err("both 'to' and 'text' are required")
}

	url := fmt.Sprintf("https://api.voyp.com/v1/messages?to=%s&text=%s", to, text)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d", resp.StatusCode))
}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse error: %v", e))
}

	return ok("message sent")
}

func HandleListConversations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	url := fmt.Sprintf("https://api.voyp.com/v1/conversations?limit=%d", limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d", resp.StatusCode))
}

	body, _ := io.ReadAll(resp.Body)
	var list []interface{}
	if e = json.Unmarshal(body, &list); e != nil {
		return err(fmt.Sprintf("parse error: %v", e))
}

	return success(fmt.Sprintf("found %d conversations", len(list)))
}