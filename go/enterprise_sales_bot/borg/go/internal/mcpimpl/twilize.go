package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func HandleSendSms_twilize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	from, _ :=getString(args, "from")
	body, _ :=getString(args, "body")
	sid, _ :=getString(args, "account_sid")
	token, _ :=getString(args, "auth_token")
	if to == "" || from == "" || body == "" || sid == "" || token == "" {
		return err("missing required fields: to, from, body, account_sid, auth_token")
}

	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", sid)
	form := url.Values{"To": {to}, "From": {from}, "Body": {body}}
	req, e := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(form.Encode()))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(sid, token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err("twilio error: " + string(bodyBytes))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	return success(fmt.Sprintf("Message sent, SID: %v", result["sid"]))
}

func HandleListMessages_twilize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sid, _ :=getString(args, "account_sid")
	token, _ :=getString(args, "auth_token")
	if sid == "" || token == "" {
		return err("missing account_sid or auth_token")
}

	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", sid)
	req, e := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(sid, token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err("twilio error: " + string(bodyBytes))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	messages, found := result["messages"].([]interface{})
	if !found {
		return err("no messages field in response")
}

	return success(fmt.Sprintf("Found %d messages", len(messages)))
}