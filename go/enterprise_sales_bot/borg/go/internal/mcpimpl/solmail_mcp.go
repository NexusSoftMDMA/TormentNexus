package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleSendMail_solmail_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if to == "" || subject == "" || body == "" {
		return err("to, subject, body are required")
}

	payload, e := json.Marshal(map[string]string{
		"to": to, "subject": subject, "body": body,
	})
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	resp, e := http.Post("https://api.solmail.com/v1/send", "application/json", strings.NewReader(string(payload)))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(bodyBytes)))
}

	return ok("email sent to " + to)
}

func HandleGetInbox(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	url := "https://api.solmail.com/v1/inbox/" + address
	resp, e := http.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	return success(string(body))
}