package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleMailboxValidator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	email, _ :=getString(args, "email")
	apiKey, _ :=getString(args, "apiKey")
	if email == "" || apiKey == "" {
		return err("email and apiKey are required")
}

	apiURL := fmt.Sprintf("https://api.mailboxvalidator.com/v1/validation/single?email=%s&key=%s", url.QueryEscape(email), url.QueryEscape(apiKey))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("failed to call API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON response: " + e.Error())
}

	if status, found := result["status"]; found && status != nil {
		if s, found := status.(string); found {
			return ok(fmt.Sprintf("Email validation status: %s", s))

	}
	return ok(fmt.Sprintf("Response: %v", result))
}
}