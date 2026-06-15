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
	apiKey, _ :=getString(args, "api_key")
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	text, _ :=getString(args, "text")

	if apiKey == "" || to == "" || subject == "" || text == "" {
		return err("missing required arguments: api_key, to, subject, text")
}

	body := map[string]interface{}{
		"to":      []map[string]string{{"email": to}},
		"subject": subject,
		"text":    text,
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request body: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://send.api.mailtrap.io/api/send", bytes.NewReader(jsonBody))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("Mailtrap API returned status %d: %s", resp.StatusCode, string(respBody)))
}

	return ok("Email sent successfully")
}