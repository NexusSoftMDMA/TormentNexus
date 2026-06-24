package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	chatID, _ :=getString(args, "chat_id")
	if chatID == "" {
		return err("chat_id is required")
}

	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	resp, e := http.DefaultClient.PostForm(apiURL, url.Values{"chat_id": {chatID}, "text": {text}})
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result struct {
		OK bool `json:"ok"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	if !result.OK {
		return err("telegram returned error")
}

	return ok("message sent")
}

func HandleGetUpdates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	offset, _ :=getInt(args, "offset")
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", token)
	if offset > 0 {
		apiURL += "?offset=" + fmt.Sprint(offset)

	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success(string(body))
}
}