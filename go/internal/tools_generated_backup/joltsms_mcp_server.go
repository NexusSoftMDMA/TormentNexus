package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSendSms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	to, _ :=getString(args, "to")
	message, _ :=getString(args, "message")
	sender, _ :=getString(args, "sender")
	if apiKey == "" || to == "" || message == "" {
		return err("api_key, to, and message are required")
}

	form := url.Values{}
	form.Set("api_key", apiKey)
	form.Set("to", to)
	form.Set("message", message)
	form.Set("sender", sender)
	resp, e := http.DefaultClient.PostForm("https://joltsms.com/api/send", form)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse response: %v", e))
}

	status, found := result["status"]
	if !found || fmt.Sprint(status) != "success" {
		return ok(fmt.Sprintf("response: %s", string(body)))
}

	return success("SMS sent successfully")
}

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	u := fmt.Sprintf("https://joltsms.com/api/balance?api_key=%s", url.QueryEscape(apiKey))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse response: %v", e))
}

	balance, found := result["balance"]
	if !found {
		return ok(fmt.Sprintf("response: %s", string(body)))
}

	return ok(fmt.Sprintf("balance: %v", balance))
}