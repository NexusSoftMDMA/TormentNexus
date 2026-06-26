package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	)

// Payload represents the JSON data sent in a webhook.
type Payload struct {
	Event     string       `json:"event"`
	DealID    int64        `json:"deal_id"`
	CompanyID int64        `json:"company_id"`
	OldState  string       `json:"old_state"`
	NewState  string       `json:"new_state"`
	Timestamp string       `json:"timestamp"`
}

// Client handles outbound webhooks.
type Client struct {
	url    string
	secret string
	client *http.Client
}

// NewClient creates a new webhook client.
func NewClient(url, secret string) *Client {
	return &Client{
		url:    url,
		secret: secret,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// DefaultClient creates a client from environment variables.
func DefaultClient() *Client {
	url := os.Getenv("OUTBOUND_WEBHOOK_URL")
	secret := os.Getenv("OUTBOUND_WEBHOOK_SECRET")
	if url == "" {
		return nil
	}
	return NewClient(url, secret)
}

// NotifyStateChange sends a webhook notification for a deal state change.
func (c *Client) NotifyStateChange(ctx context.Context, dealID int64, companyID int64, oldState string, newState string) error {
	if c == nil || c.url == "" {
		return nil // Webhooks not configured
	}

	payload := Payload{
		Event:     "deal.state_changed",
		DealID:    dealID,
		CompanyID: companyID,
		OldState:  oldState,
		NewState:  newState,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add HMAC signature if secret is provided
	if c.secret != "" {
		mac := hmac.New(sha256.New, []byte(c.secret))
		mac.Write(body)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Webhook-Signature-256", "sha256="+signature)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	slog.InfoContext(ctx, "Webhook sent successfully", "event", payload.Event, "deal_id", dealID)
	return nil
}
