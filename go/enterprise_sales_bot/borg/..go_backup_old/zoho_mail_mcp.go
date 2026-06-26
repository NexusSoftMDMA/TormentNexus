package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	accessToken, _ :=getString(args, "accessToken")
	if to == "" || subject == "" || body == "" || accessToken == "" {
		return err("Missing required fields: to, subject, body, accessToken")
}

	payload := map[string]string{"to": to, "subject": subject, "body": body}
	data, e := json.Marshal(payload)
	if e != nil {
		return err("Failed to marshal payload: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://mail.zoho.com/api/accounts/me/messages", bytes.NewBuffer(data))
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	bodyBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Zoho API error %d: %s", resp.StatusCode, string(bodyBytes)))
}

	return ok("Email sent successfully")
}

func HandleListEmails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	folderID, _ :=getString(args, "folderID")
	accessToken, _ :=getString(args, "accessToken")
	if accessToken == "" {
		return err("Missing required field: accessToken")
}

	url := "https://mail.zoho.com/api/accounts/me/messages"
	if folderID != "" {
		url += "?folderId=" + folderID
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	bodyBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Zoho API error %d: %s", resp.StatusCode, string(bodyBytes)))
}

	return success(string(bodyBytes))
}