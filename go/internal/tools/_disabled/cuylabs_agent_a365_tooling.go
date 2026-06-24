package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func HandleSendMail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"subject": subject,
			"toRecipients": []map[string]interface{}{
				{"emailAddress": map[string]string{"address": to}},
			},
			"body": map[string]string{"contentType": "Text", "content": body},
		},
		"saveToSentItems": true,
	}
	data, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
}

	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", os.Getenv("AGENT365_USER_ID"))
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("AGENT365_ACCESS_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := ioutil.ReadAll(resp.Body)
		return err("API error: " + string(b))
}

	return ok("email sent")
}

func HandleCreateCalendarEvent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subject, _ :=getString(args, "subject")
	start, _ :=getString(args, "start")
	end, _ :=getString(args, "end")
	attendees, _ :=getString(args, "attendees")
	payload := map[string]interface{}{
		"subject": subject,
		"start":   map[string]string{"dateTime": start, "timeZone": "UTC"},
		"end":     map[string]string{"dateTime": end, "timeZone": "UTC"},
		"attendees": []map[string]interface{}{
			{"emailAddress": map[string]string{"address": attendees}},
		},
	}
	data, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
}

	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/events", os.Getenv("AGENT365_USER_ID"))
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("AGENT365_ACCESS_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := ioutil.ReadAll(resp.Body)
		return err("API error: " + string(b))
}

	return ok("event created")
}