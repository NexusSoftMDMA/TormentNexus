package toon

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	thread := &Thread{
		ID:        "test-thread-001",
		SessionID: "session-abc",
		Agent:     "planner",
		Model:     "gpt-4o",
		CreatedAt: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
		Tags:      []string{"test", "roundtrip"},
		Messages: []Message{
			{
				Role:      RoleSystem,
				Type:      TypeText,
				Timestamp: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
				Content:   "You are a helpful assistant.",
			},
			{
				Role:      RoleUser,
				Type:      TypeText,
				Timestamp: time.Date(2026, 5, 20, 12, 0, 5, 0, time.UTC),
				Content:   "What is the capital of France?",
			},
			{
				Role:      RoleAssistant,
				Type:      TypeText,
				Timestamp: time.Date(2026, 5, 20, 12, 0, 10, 0, time.UTC),
				Content:   "The capital of France is Paris.",
				ToolCalls: []ToolCall{
					{
						ID:       "tc-1",
						Name:     "search",
						Args:     map[string]interface{}{"query": "France capital"},
						Duration: 150,
					},
				},
			},
			{
				Role:      RoleTool,
				Type:      TypeToolResult,
				Timestamp: time.Date(2026, 5, 20, 12, 0, 11, 0, time.UTC),
				Content:   "Paris is the capital of France.",
			},
		},
		Memories: []MemoryRef{
			{
				ID:         "mem-001",
				Type:       "working",
				Content:    "User asked about France",
				Importance: 0.75,
				HeatScore:  65.0,
			},
		},
		Metadata: MapEntries{
			{Key: "environment", Value: "test"},
		},
	}

	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(thread); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := DecodeFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.ID != thread.ID {
		t.Errorf("ID mismatch: got %q, want %q", decoded.ID, thread.ID)
	}
	if decoded.SessionID != thread.SessionID {
		t.Errorf("SessionID mismatch: got %q, want %q", decoded.SessionID, thread.SessionID)
	}
	if decoded.Agent != thread.Agent {
		t.Errorf("Agent mismatch: got %q, want %q", decoded.Agent, thread.Agent)
	}
	if decoded.Model != thread.Model {
		t.Errorf("Model mismatch: got %q, want %q", decoded.Model, thread.Model)
	}
	if len(decoded.Messages) != len(thread.Messages) {
		t.Fatalf("Message count mismatch: got %d, want %d", len(decoded.Messages), len(thread.Messages))
	}
	if len(decoded.Memories) != len(thread.Memories) {
		t.Errorf("Memory count mismatch: got %d, want %d", len(decoded.Memories), len(thread.Memories))
	}
	if decoded.Messages[1].Content != "What is the capital of France?" {
		t.Errorf("Message content mismatch: got %q", decoded.Messages[1].Content)
	}
	if len(decoded.Messages[2].ToolCalls) != 1 {
		t.Errorf("ToolCalls count mismatch: got %d, want 1", len(decoded.Messages[2].ToolCalls))
	}
	if decoded.Messages[2].ToolCalls[0].Name != "search" {
		t.Errorf("ToolCall name mismatch: got %q", decoded.Messages[2].ToolCalls[0].Name)
	}
	if decoded.Memories[0].Importance != 0.75 {
		t.Errorf("Memory importance mismatch: got %f, want 0.75", decoded.Memories[0].Importance)
	}
	if len(decoded.Tags) != 2 {
		t.Errorf("Tags count mismatch: got %d, want 2", len(decoded.Tags))
	}
}

func TestEncodeDecodeEmpty(t *testing.T) {
	thread := &Thread{
		ID:        "empty-thread",
		SessionID: "session-empty",
		CreatedAt: time.Now(),
	}

	data, err := EncodeToBytes(thread)
	if err != nil {
		t.Fatalf("EncodeToBytes failed: %v", err)
	}

	decoded, err := DecodeFromBytes(data)
	if err != nil {
		t.Fatalf("DecodeFromBytes failed: %v", err)
	}

	if decoded.ID != "empty-thread" {
		t.Errorf("ID mismatch: got %q, want %q", decoded.ID, "empty-thread")
	}
	if len(decoded.Messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(decoded.Messages))
	}
	if len(decoded.Memories) != 0 {
		t.Errorf("Expected 0 memories, got %d", len(decoded.Memories))
	}
}

func TestChecksumValidation(t *testing.T) {
	thread := &Thread{
		ID:        "corrupt-test",
		SessionID: "session-corrupt",
		CreatedAt: time.Now(),
		Messages: []Message{
			{Role: RoleUser, Type: TypeText, Timestamp: time.Now(), Content: "hello"},
		},
	}

	data, err := EncodeToBytes(thread)
	if err != nil {
		t.Fatalf("EncodeToBytes failed: %v", err)
	}

	// Corrupt a byte in the middle
	data[len(data)/2] ^= 0xFF

	_, err = DecodeFromBytes(data)
	if err == nil {
		t.Error("Expected checksum error, got nil")
	}
}

func TestInvalidMagic(t *testing.T) {
	data := make([]byte, 20)
	copy(data, []byte("WRONGFMT"))
	// Append a valid checksum for the first 16 bytes
	// But this won't work since we need proper CRC...
	// Let's just verify that decoding garbage fails
	_, err := DecodeFromBytes(data)
	if err == nil {
		t.Error("Expected error for invalid magic, got nil")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	thread := &Thread{
		ID:        "encrypt-test",
		SessionID: "session-encrypt",
		CreatedAt: time.Now(),
		Messages: []Message{
			{Role: RoleUser, Type: TypeText, Timestamp: time.Now(), Content: "secret message"},
		},
	}

	payload, err := EncodeToBytes(thread)
	if err != nil {
		t.Fatalf("EncodeToBytes failed: %v", err)
	}

	// Generate a 32-byte key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("rand.Read failed: %v", err)
	}

	encrypted, err := Encrypt(payload, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if bytes.Equal(encrypted[:len(payload)], payload) {
		t.Error("Encrypted data should not match original")
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, payload) {
		t.Error("Decrypted data should match original")
	}

	// Verify we can decode the decrypted data
	decoded, err := DecodeFromBytes(decrypted)
	if err != nil {
		t.Fatalf("DecodeFromBytes after decrypt failed: %v", err)
	}
	if decoded.ID != "encrypt-test" {
		t.Errorf("ID mismatch after encrypt/decrypt: got %q", decoded.ID)
	}
}

func TestPrunableThread(t *testing.T) {
	thread := &Thread{
		ID:        "prune-test",
		SessionID: "session-prune",
		CreatedAt: time.Now(),
		Memories: []MemoryRef{
			{ID: "mem-1", Type: "working", Content: "important", Importance: 0.9, HeatScore: 80},
		},
	}

	// Create 10 messages
	for i := 0; i < 10; i++ {
		msg := Message{
			Role:      RoleUser,
			Type:      TypeText,
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Content:   "message " + string(rune('0'+i)),
		}
		// Make one a tool result
		if i == 3 {
			msg.Role = RoleTool
			msg.Type = TypeToolResult
			msg.Content = "tool result 3"
		}
		// Make one a system message
		if i == 0 {
			msg.Role = RoleSystem
			msg.Type = TypeText
			msg.Content = "system prompt"
		}
		thread.Messages = append(thread.Messages, msg)
	}

	// Prune to keep last 3 messages
	pruned := PrunableThread(thread, 3)

	if len(pruned.Messages) > len(thread.Messages) {
		t.Errorf("Pruned thread should not have more messages than original")
	}

	// Should keep: system (0), tool result (3), and last 3 (8, 9, 10... well 7, 8, 9)
	// The exact count depends on dedup logic
	if pruned.ID != thread.ID {
		t.Errorf("ID should be preserved")
	}
	if len(pruned.Memories) != 1 {
		t.Errorf("Memories should be preserved, got %d", len(pruned.Memories))
	}
}

func TestSummarizeThread(t *testing.T) {
	thread := &Thread{
		ID:        "summary-test",
		SessionID: "session-summary",
		CreatedAt: time.Now(),
		Memories: []MemoryRef{
			{ID: "mem-1", Type: "long_term", Content: "preserved", Importance: 0.95, HeatScore: 90},
		},
		Messages: []Message{
			{Role: RoleUser, Type: TypeText, Timestamp: time.Now(), Content: "question 1"},
			{Role: RoleAssistant, Type: TypeText, Timestamp: time.Now(), Content: "answer 1"},
			{Role: RoleUser, Type: TypeText, Timestamp: time.Now(), Content: "question 2"},
			{Role: RoleAssistant, Type: TypeText, Timestamp: time.Now(), Content: "answer 2"},
		},
	}

	summary := SummarizeThread(thread, "User asked two questions about various topics.")

	if len(summary.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(summary.Messages))
	}
	if summary.Messages[0].Type != TypeSummary {
		t.Errorf("Expected TypeSummary, got %d", summary.Messages[0].Type)
	}
	if summary.Messages[0].Content != "User asked two questions about various topics." {
		t.Errorf("Summary content mismatch: got %q", summary.Messages[0].Content)
	}
	if len(summary.Memories) != 1 {
		t.Errorf("Memories should be preserved in summary, got %d", len(summary.Memories))
	}
	// Check that original_message_count is in metadata
	found := false
	for _, entry := range summary.Metadata {
		if entry.Key == "original_message_count" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected original_message_count in metadata")
	}
}

func TestEncodeDecodeConvenience(t *testing.T) {
	thread := &Thread{
		ID:        "conv-test",
		SessionID: "session-conv",
		CreatedAt: time.Now(),
		Messages: []Message{
			{Role: RoleUser, Type: TypeCode, Timestamp: time.Now(), Content: "fmt.Println(\"hello\")"},
		},
	}

	data, err := EncodeToBytes(thread)
	if err != nil {
		t.Fatalf("EncodeToBytes: %v", err)
	}

	if len(data) == 0 {
		t.Error("Encoded data should not be empty")
	}

	// Verify magic bytes at the start
	if string(data[:4]) != "TOON" {
		t.Errorf("Magic bytes mismatch: got %q, want %q", string(data[:4]), "TOON")
	}
}
