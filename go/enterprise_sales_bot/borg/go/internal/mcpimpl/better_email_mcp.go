package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleSendEmail_better_email_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if to == "" || subject == "" || body == "" {
		return err("to, subject, and body are required")
}

	payload := fmt.Sprintf(`{"to":"%s","subject":"%s","body":"%s"}`, to, subject, body)
	resp, e := http.DefaultClient.Post("https://api.sendgrid.com/v3/mail/send", "application/json", strings.NewReader(payload))
	if e != nil {
		return err("failed to send email: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 202 {
		return err("email API returned status " + resp.Status)
}

	return ok("email sent successfully")
}

func HandleListInbox_better_email_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	maxResults, _ :=getInt(args, "max_results")
	if maxResults <= 0 {
		maxResults = 10
	}
	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.sendgrid.com/v3/messages?limit=%d", maxResults))
	if e != nil {
		return err("failed to fetch inbox: " + e.Error())
}

	defer resp.Body.Close()
	var messages []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&messages); e != nil {
		return err("failed to decode inbox response: " + e.Error())
}

	return ok(fmt.Sprintf("found %d messages", len(messages)))
}