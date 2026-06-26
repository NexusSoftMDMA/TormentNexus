package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleListMessages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	query, _ :=getString(args, "query")
	if query == "" {
		query = "in:inbox"
	}
	url := fmt.Sprintf("https://gmail.googleapis.com/gmail/v1/users/me/messages?q=%s", query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	messages, found := result["messages"].([]interface{})
	if !found {
		return ok("No messages found.")
}

	var ids []string
	for _, m := range messages {
		msg, _ := m.(map[string]interface{})
		if id, found := msg["id"].(string); found {
			ids = append(ids, id)

	}
	return ok(fmt.Sprintf("Found %d messages: %s", len(ids), strings.Join(ids, ", ")))
}

}

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if to == "" || subject == "" {
		return err("to and subject are required")
}

	rawEmail := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)
	encoded := base64URLEncode([]byte(rawEmail))
	payload := map[string]string{"raw": encoded}
	jsonPayload, e := json.Marshal(payload)
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	url := "https://gmail.googleapis.com/gmail/v1/users/me/messages/send"
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonPayload)))
	if e != nil {
		return err("create request failed: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("send failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	return ok("Email sent successfully.")
}

// base64URLEncode is a helper to encode raw email in base64 URL-safe format (without padding)
func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}