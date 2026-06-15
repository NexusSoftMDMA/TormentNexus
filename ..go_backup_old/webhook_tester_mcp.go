package tools

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateWebhook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	b := make([]byte, 16)
	_, e := rand.Read(b)
	if e != nil {
		return err("failed to generate webhook ID")
}

	id := hex.EncodeToString(b)
	url := fmt.Sprintf("https://webhook.site/%s", id)
	return ok(fmt.Sprintf("Webhook endpoint created: %s", url))
}

func HandleTriggerWebhook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	payload, _ :=getString(args, "payload")
	if payload == "" {
		payload = "{\"test\":true}"
	}
	var body bytes.Buffer
	e := json.NewEncoder(&body).Encode(map[string]interface{}{"payload": payload})
	if e != nil {
		return err("invalid payload")
}

	resp, e := http.DefaultClient.Post(url, "application/json", &body)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(fmt.Sprintf("Status: %d, Response: %s", resp.StatusCode, string(data)))
}