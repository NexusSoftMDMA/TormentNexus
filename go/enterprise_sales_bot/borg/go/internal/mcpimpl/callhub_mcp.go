package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func HandleCallhubSendSms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	phone, _ :=getString(args, "phone")
	message, _ :=getString(args, "message")
	if phone == "" || message == "" {
		return err("phone and message are required")
}

	body := fmt.Sprintf(`{"to":"%s","text":"%s"}`, phone, message)
	req, e := http.NewRequestWithContext(ctx, "POST", os.Getenv("CALLHUB_API_URL")+"/sms", strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("CALLHUB_API_KEY"))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("API returned error status: " + resp.Status)
}

	return ok("SMS sent successfully")
}

func HandleCallhubGetContacts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", os.Getenv("CALLHUB_API_URL")+"/contacts", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("CALLHUB_API_KEY"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("API returned error status: " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(fmt.Sprintf("Contacts: %+v", result))
}