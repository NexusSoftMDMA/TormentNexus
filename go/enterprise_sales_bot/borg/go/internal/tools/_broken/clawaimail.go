package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func HandleSendEmail_clawaimail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if to == "" {
		return err("missing 'to' recipient")
}

	payload := fmt.Sprintf(`{"to":%q,"subject":%q,"body":%q}`, to, subject, body)
	apiURL := os.Getenv("CLAWAMAIL_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	req, e := http.NewRequestWithContext(ctx, "POST", apiURL+"/send", strings.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("send request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("server error %d: %s", resp.StatusCode, string(bodyBytes)))
}

	return ok("email sent successfully")
}

func HandleListEmails_clawaimail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	folder, _ :=getString(args, "folder")
	if folder == "" {
		folder = "INBOX"
	}
	apiURL := os.Getenv("CLAWAMAIL_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", apiURL+"/list?folder="+folder, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("list request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}