package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	)

func TestNotifyStateChange(t *testing.T) {
	var receivedPayload Payload
	var receivedSignature string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedPayload)
		receivedSignature = r.Header.Get("X-Webhook-Signature-256")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "mysecret")

	err := client.NotifyStateChange(context.Background(), 1, 10, "Discovered", "Researched")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if receivedPayload.DealID != 1 || receivedPayload.OldState != "Discovered" || receivedPayload.NewState != "Researched" {
		t.Errorf("unexpected payload: %+v", receivedPayload)
	}

	if receivedSignature == "" {
		t.Error("expected signature header, got empty")
	}
}

func TestNotifyStateChange_NilClient(t *testing.T) {
	var client *Client
	err := client.NotifyStateChange(context.Background(), 0, 0, "", "")
	if err != nil {
		t.Fatalf("expected no error for nil client, got %v", err)
	}
}
