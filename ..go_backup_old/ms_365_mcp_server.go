package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListUsers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/users", nil)
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
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	return success(string(body))
}

func HandleSendMail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	userID, _ :=getString(args, "user_id")
	recipient, _ :=getString(args, "recipient")
	subject, _ :=getString(args, "subject")
	bodyContent, _ :=getString(args, "body")
	if token == "" || userID == "" || recipient == "" || subject == "" || bodyContent == "" {
		return err("token, user_id, recipient, subject, body required")
}

	msg := map[string]interface{}{
		"message": map[string]interface{}{
			"subject": subject,
			"body": map[string]interface{}{
				"contentType": "Text",
				"content":     bodyContent,
			},
			"toRecipients": []interface{}{
				map[string]interface{}{
					"emailAddress": map[string]interface{}{
						"address": recipient,
					},
				},
			},
		},
	}
	payload, e := json.Marshal(msg)
	if e != nil {
		return err("json marshal failed: " + e.Error())
}

	url := "https://graph.microsoft.com/v1.0/users/" + userID + "/sendMail"
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return err("API error: " + string(body))
}

	return ok("Email sent successfully")
}