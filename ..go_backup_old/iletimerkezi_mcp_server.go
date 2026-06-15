package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSendSms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	sender, _ :=getString(args, "sender")
	phone, _ :=getString(args, "phone")
	message, _ :=getString(args, "message")
	url := fmt.Sprintf("https://api.iletimerkezi.com/v1/send-sms?key=%s&sender=%s&phone=%s&message=%s", apiKey, sender, phone, message)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to send sms: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid json response")
}

	status, found := result["status"].(string)
	if !found || status != "success" {
		return err("api returned non-success status")
}

	return ok("SMS sent successfully")
}

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	url := fmt.Sprintf("https://api.iletimerkezi.com/v1/get-balance?key=%s", apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get balance: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid json response")
}

	balance, found := data["balance"].(float64)
	if !found {
		return err("balance field missing")
}

	return success(fmt.Sprintf("Current balance: %.2f credits", balance))
}