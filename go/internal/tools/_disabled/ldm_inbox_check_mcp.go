package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListInbox(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	url := fmt.Sprintf("https://api.ldm.example.com/inbox?limit=%d", limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch inbox: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Inbox messages: %v", result))
}

func HandleGetMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	messageID, _ :=getString(args, "messageId")
	if messageID == "" {
		return err("messageId is required")
}

	url := fmt.Sprintf("https://api.ldm.example.com/messages/%s", messageID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch message: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Message: %v", result))
}