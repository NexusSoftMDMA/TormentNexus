package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleListEmails_outlook_assistant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "accessToken")
	folder, _ :=getString(args, "folder")
	if folder == "" {
		folder = "inbox"
	}
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/mailFolders/%s/messages", folder)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Emails: %v", result))
}

func HandleSendEmail_outlook_assistant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "accessToken")
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if to == "" || subject == "" {
		return err("to and subject are required")
}

	payload := fmt.Sprintf(`{"message":{"subject":"%s","body":{"contentType":"Text","content":"%s"},"toRecipients":[{"emailAddress":{"address":"%s"}}]},"saveToSentItems":true}`, subject, body, to)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://graph.microsoft.com/v1.0/me/sendMail", strings.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("send failed: %s", string(bodyBytes)))
}

	return success("Email sent successfully")
}