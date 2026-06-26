package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSendMarketingEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if to == "" || subject == "" || body == "" {
		return err("missing required fields: to, subject, body")
}

	payload := map[string]string{"to": to, "subject": subject, "body": body}
	data, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	resp, e := http.DefaultClient.Post("https://api.bolide.ai/send", "application/json", io.NopCloser(bytes.NewReader(data)))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("bad status: %d", resp.StatusCode))
}

	return ok("email sent successfully")
}

func HandleTrackCampaign(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	campaignID, _ :=getString(args, "campaign_id")
	if campaignID == "" {
		return err("missing campaign_id")
}

	url := fmt.Sprintf("https://api.bolide.ai/track?campaign_id=%s", campaignID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	return ok(string(body))
}