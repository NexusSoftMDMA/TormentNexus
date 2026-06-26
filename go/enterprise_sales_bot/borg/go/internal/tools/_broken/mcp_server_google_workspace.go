package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGmailSend(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "accessToken")
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if to == "" || subject == "" {
		return err("missing required fields: to, subject")
}

	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)
	raw := map[string]string{"raw": encodeBase64([]byte(msg))}
	payload, _ := json.Marshal(raw)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://gmail.googleapis.com/gmail/v1/users/me/messages/send", bytes.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("gmail API error: " + resp.Status)
}

	return ok("email sent successfully")
}

func HandleCalendarCreate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "accessToken")
	summary, _ :=getString(args, "summary")
	start, _ :=getString(args, "startDateTime")
	end, _ :=getString(args, "endDateTime")
	if summary == "" || start == "" || end == "" {
		return err("missing required fields: summary, startDateTime, endDateTime")
}

	event := map[string]interface{}{
		"summary": summary,
		"start":   map[string]string{"dateTime": start, "timeZone": "UTC"},
		"end":     map[string]string{"dateTime": end, "timeZone": "UTC"},
	}
	payload, _ := json.Marshal(event)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://www.googleapis.com/calendar/v3/calendars/primary/events", bytes.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("calendar API error: " + resp.Status)
}

	return ok("event created successfully")
}

func encodeBase64(data []byte) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result []byte
	for i := 0; i < len(data); i += 3 {
		var val int
		remaining := len(data) - i
		if remaining >= 3 {
			val = int(data[i])<<16 | int(data[i+1])<<8 | int(data[i+2])
		} else if remaining == 2 {
			val = int(data[i])<<16 | int(data[i+1])<<8
		} else {
			val = int(data[i]) << 16
		}
		result = append(result, base64Chars[(val>>18)&0x3F], base64Chars[(val>>12)&0x3F])
		if remaining >= 3 {
			result = append(result, base64Chars[(val>>6)&0x3F], base64Chars[val&0x3F])
		} else if remaining == 2 {
			result = append(result, base64Chars[(val>>6)&0x3F], '=')
		} else {
			result = append(result, '=', '=')

	}
	return string(result)
}
}