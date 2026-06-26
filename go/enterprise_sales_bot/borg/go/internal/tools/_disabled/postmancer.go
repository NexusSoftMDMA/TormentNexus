package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	textBody, _ :=getString(args, "textBody")
	apiKey, _ :=getString(args, "apiKey")
	if to == "" || subject == "" || textBody == "" || apiKey == "" {
		return err("missing required fields: to, subject, textBody, apiKey")
}

	payload := fmt.Sprintf(`{"From":"sender@example.com","To":"%s","Subject":"%s","TextBody":"%s"}`, to, subject, textBody)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.postmarkapp.com/email", strings.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send email: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode >= 300 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Email sent! MessageID: %v", result["MessageID"]))
}