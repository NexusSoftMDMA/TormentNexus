package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"

	"golang.org/x/crypto/nacl/box"
)

func TestNewArchiveFromConfig_InvalidNaClKey(t *testing.T) {
	dispatcher := &ArchiveDispatcher{} // minimal, just needs non-nil

	tests := []struct {
		name string
		key  string
	}{
		{"bad base64", "not-valid!!!"},
		{"wrong length", base64.StdEncoding.EncodeToString([]byte("tooshort"))},
		{"truncated key", base64.StdEncoding.EncodeToString(make([]byte, 16))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ArchiveConfig{
				URL:              "https://example.com/archive",
				Include:          "both",
				NaClRecipientKey: tt.key,
			}
			cfgJSON, _ := json.Marshal(cfg)
			_, err := NewArchiveFromConfig(cfgJSON, nil, dispatcher)
			if err == nil {
				t.Fatalf("expected error for key %q, got nil", tt.key)
			}
		})
	}
}

func TestNewArchiveFromConfig_ValidNaClKey(t *testing.T) {
	dispatcher := &ArchiveDispatcher{}
	pub, _, _ := box.GenerateKey(rand.Reader)
	b64Key := base64.StdEncoding.EncodeToString(pub[:])

	cfg := ArchiveConfig{
		URL:              "https://example.com/archive",
		Include:          "both",
		NaClRecipientKey: b64Key,
	}
	cfgJSON, _ := json.Marshal(cfg)
	mw, err := NewArchiveFromConfig(cfgJSON, nil, dispatcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	archive := mw.(*Archive)
	if archive.recipientKey == nil {
		t.Fatal("recipientKey should be cached, got nil")
	}
	if *archive.recipientKey != *pub {
		t.Error("cached key does not match original")
	}
}

func TestNewArchiveFromConfig_NoEncryption(t *testing.T) {
	dispatcher := &ArchiveDispatcher{}

	cfg := ArchiveConfig{
		URL:     "https://example.com/archive",
		Include: "both",
	}
	cfgJSON, _ := json.Marshal(cfg)
	mw, err := NewArchiveFromConfig(cfgJSON, nil, dispatcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	archive := mw.(*Archive)
	if archive.recipientKey != nil {
		t.Error("recipientKey should be nil when encryption is disabled")
	}
}

func TestValidateArchiveConfig(t *testing.T) {
	pub, _, _ := box.GenerateKey(rand.Reader)
	goodKey := base64.StdEncoding.EncodeToString(pub[:])

	tests := []struct {
		name    string
		cfg     ArchiveConfig
		wantErr bool
	}{
		{
			name:    "missing url",
			cfg:     ArchiveConfig{Include: "both"},
			wantErr: true,
		},
		{
			name:    "plain http on public host",
			cfg:     ArchiveConfig{URL: "http://example.com/archive", Include: "both"},
			wantErr: true,
		},
		{
			name:    "plain http on localhost allowed",
			cfg:     ArchiveConfig{URL: "http://localhost:4000/ingest", Include: "both"},
			wantErr: false,
		},
		{
			name:    "plain http on 127.0.0.1 allowed",
			cfg:     ArchiveConfig{URL: "http://127.0.0.1:4000/ingest", Include: "both"},
			wantErr: false,
		},
		{
			name:    "https ok",
			cfg:     ArchiveConfig{URL: "https://example.com/archive", Include: "both"},
			wantErr: false,
		},
		{
			name:    "unknown scheme rejected",
			cfg:     ArchiveConfig{URL: "ftp://example.com/archive", Include: "both"},
			wantErr: true,
		},
		{
			name:    "unknown auth_type rejected",
			cfg:     ArchiveConfig{URL: "https://example.com/archive", AuthType: "hmac"},
			wantErr: true,
		},
		{
			name:    "bearer with empty value rejected",
			cfg:     ArchiveConfig{URL: "https://example.com/archive", AuthType: "bearer", AuthValue: ""},
			wantErr: true,
		},
		{
			name:    "bearer with value accepted",
			cfg:     ArchiveConfig{URL: "https://example.com/archive", AuthType: "bearer", AuthValue: "tok"},
			wantErr: false,
		},
		{
			name:    "api_key with empty value rejected",
			cfg:     ArchiveConfig{URL: "https://example.com/archive", AuthType: "api_key", AuthValue: ""},
			wantErr: true,
		},
		{
			name:    "0.0.0.0 is not loopback",
			cfg:     ArchiveConfig{URL: "http://0.0.0.0:4000/ingest"},
			wantErr: true,
		},
		{
			name:    "IPv6 loopback literal allowed",
			cfg:     ArchiveConfig{URL: "http://[::1]:4000/ingest"},
			wantErr: false,
		},
		{
			name:    "bad include rejected",
			cfg:     ArchiveConfig{URL: "https://example.com/archive", Include: "everything"},
			wantErr: true,
		},
		{
			name:    "bad nacl key rejected",
			cfg:     ArchiveConfig{URL: "https://example.com/archive", NaClRecipientKey: "not-valid!!!"},
			wantErr: true,
		},
		{
			name:    "good nacl key accepted",
			cfg:     ArchiveConfig{URL: "https://example.com/archive", NaClRecipientKey: goodKey},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateArchiveConfig(tt.cfg)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateArchiveConfig_ReturnsDecodedKey(t *testing.T) {
	pub, _, _ := box.GenerateKey(rand.Reader)
	b64 := base64.StdEncoding.EncodeToString(pub[:])
	cfg := ArchiveConfig{URL: "https://example.com/archive", NaClRecipientKey: b64}
	decoded, err := ValidateArchiveConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decoded == nil {
		t.Fatal("decoded key should not be nil when NaClRecipientKey is set")
	}
	if *decoded != *pub {
		t.Error("decoded key does not match configured pubkey")
	}
}

func TestArchive_CachedKeyUsedForEncryption(t *testing.T) {
	// Verify that the cached recipientKey matches what was configured
	dispatcher := &ArchiveDispatcher{}

	pub, _, _ := box.GenerateKey(rand.Reader)
	b64Key := base64.StdEncoding.EncodeToString(pub[:])

	cfg := ArchiveConfig{
		URL:              "https://example.com/archive",
		Include:          "both",
		NaClRecipientKey: b64Key,
	}
	cfgJSON, _ := json.Marshal(cfg)
	mw, err := NewArchiveFromConfig(cfgJSON, nil, dispatcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	archive := mw.(*Archive)

	// The cached key should allow encryption without re-decoding
	payload := []byte(`{"test": true}`)
	encrypted, err := encryptPayload(payload, *archive.recipientKey)
	if err != nil {
		t.Fatalf("encryption with cached key failed: %v", err)
	}
	if len(encrypted) == 0 {
		t.Error("encrypted payload is empty")
	}
}
