package mcpimpl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleSlackTools(t *testing.T) {
	// Start a local mock Slack API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Auth check
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"ok": false, "error": "not_authed"}`))
			return
		}

		path := r.URL.Path
		switch path {
		case "/conversations.list":
			w.Write([]byte(`{"ok": true, "channels": [{"id": "C1", "name": "general"}]}`))
		case "/conversations.info":
			w.Write([]byte(`{"ok": true, "channel": {"id": "C1", "name": "general", "is_archived": false}}`))
		case "/chat.postMessage":
			w.Write([]byte(`{"ok": true, "message": {"text": "hello"}}`))
		case "/reactions.add":
			w.Write([]byte(`{"ok": true}`))
		case "/conversations.history":
			w.Write([]byte(`{"ok": true, "messages": [{"text": "msg1"}]}`))
		case "/conversations.replies":
			w.Write([]byte(`{"ok": true, "messages": [{"text": "reply1"}]}`))
		case "/users.list":
			w.Write([]byte(`{"ok": true, "members": [{"id": "U1", "name": "alice"}]}`))
		case "/users.profile.get":
			w.Write([]byte(`{"ok": true, "profile": {"real_name": "Alice"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Setup test environment variables
	os.Setenv("SLACK_BOT_TOKEN", "test-token")
	os.Setenv("SLACK_API_URL", server.URL+"/")
	defer os.Unsetenv("SLACK_BOT_TOKEN")
	defer os.Unsetenv("SLACK_API_URL")

	ctx := context.Background()

	// Test 1: HandleSlackListChannels
	resp, err := HandleSlackListChannels(ctx, map[string]interface{}{"limit": 10.0})
	if err != nil {
		t.Fatalf("HandleSlackListChannels failed: %v", err)
	}
	if resp.IsError {
		t.Errorf("HandleSlackListChannels returned error: %s", resp.Content[0].Text)
	}
	if !strings.Contains(resp.Content[0].Text, "general") {
		t.Errorf("Expected general in channels list response, got: %s", resp.Content[0].Text)
	}

	// Test 1b: HandleSlackListChannels with SLACK_CHANNEL_IDS
	os.Setenv("SLACK_CHANNEL_IDS", "C1")
	defer os.Unsetenv("SLACK_CHANNEL_IDS")
	respPre, errPre := HandleSlackListChannels(ctx, map[string]interface{}{})
	if errPre != nil {
		t.Fatalf("HandleSlackListChannels predefined failed: %v", errPre)
	}
	if !strings.Contains(respPre.Content[0].Text, "general") {
		t.Errorf("Expected general in predefined channels list response, got: %s", respPre.Content[0].Text)
	}
	os.Unsetenv("SLACK_CHANNEL_IDS")

	// Test 2: HandleSlackPostMessage
	resp, err = HandleSlackPostMessage(ctx, map[string]interface{}{
		"channel_id": "C1",
		"text":       "hello",
	})
	if err != nil {
		t.Fatalf("HandleSlackPostMessage failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "hello") {
		t.Errorf("Expected hello in postMessage response, got: %s", resp.Content[0].Text)
	}

	// Test 3: HandleSlackReplyToThread
	resp, err = HandleSlackReplyToThread(ctx, map[string]interface{}{
		"channel_id": "C1",
		"thread_ts":  "1234567890.123456",
		"text":       "reply text",
	})
	if err != nil {
		t.Fatalf("HandleSlackReplyToThread failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "hello") {
		t.Errorf("Expected response from thread reply, got: %s", resp.Content[0].Text)
	}

	// Test 4: HandleSlackAddReaction
	resp, err = HandleSlackAddReaction(ctx, map[string]interface{}{
		"channel_id": "C1",
		"timestamp":  "1234567890.123456",
		"reaction":   "thumbsup",
	})
	if err != nil {
		t.Fatalf("HandleSlackAddReaction failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "ok") {
		t.Errorf("Expected ok response from reaction add, got: %s", resp.Content[0].Text)
	}

	// Test 5: HandleSlackGetChannelHistory
	resp, err = HandleSlackGetChannelHistory(ctx, map[string]interface{}{
		"channel_id": "C1",
		"limit":      5.0,
	})
	if err != nil {
		t.Fatalf("HandleSlackGetChannelHistory failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "msg1") {
		t.Errorf("Expected msg1 in history, got: %s", resp.Content[0].Text)
	}

	// Test 6: HandleSlackGetThreadReplies
	resp, err = HandleSlackGetThreadReplies(ctx, map[string]interface{}{
		"channel_id": "C1",
		"thread_ts":  "1234567890.123456",
	})
	if err != nil {
		t.Fatalf("HandleSlackGetThreadReplies failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "reply1") {
		t.Errorf("Expected reply1 in thread replies, got: %s", resp.Content[0].Text)
	}

	// Test 7: HandleSlackGetUsers
	resp, err = HandleSlackGetUsers(ctx, map[string]interface{}{
		"limit": 10.0,
	})
	if err != nil {
		t.Fatalf("HandleSlackGetUsers failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "alice") {
		t.Errorf("Expected alice in users list, got: %s", resp.Content[0].Text)
	}

	// Test 8: HandleSlackGetUserProfile
	resp, err = HandleSlackGetUserProfile(ctx, map[string]interface{}{
		"user_id": "U1",
	})
	if err != nil {
		t.Fatalf("HandleSlackGetUserProfile failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "Alice") {
		t.Errorf("Expected Alice in user profile, got: %s", resp.Content[0].Text)
	}
}
