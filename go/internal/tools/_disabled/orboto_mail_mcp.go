package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleSend(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if to == "" || subject == "" {
		return err("missing required fields: to, subject")
}

	payload := map[string]string{"to": to, "subject": subject, "body": body}
	data, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal json")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.orboto.com/mail/send", bytes.NewReader(data))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OMS_API_KEY"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("api error: %d", resp.StatusCode))
}

	return success("email sent to " + to)
}

func HandleQuota(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.orboto.com/mail/quota", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("OMS_API_KEY"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("api error: %d", resp.StatusCode))
}

	return success("quota check successful")
}// touch 1781132137
