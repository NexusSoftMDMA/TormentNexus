package license

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestVerifyLicense(t *testing.T) {
	privateKeyHex := "a0dacdd3763565c8d5340fab79956273c71fbde52ea09a8d8b737d41fcff9eca9a9d5d9cc7acebbbf80adfe9005586c3f6496e82e7fa300920b831397c1cb763"
	
	privKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		t.Fatalf("failed to decode private key: %v", err)
	}
	privKey := ed25519.PrivateKey(privKeyBytes)

	tempDir, err := os.MkdirTemp("", "lic-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	holder := "Test Company"
	seats := 5
	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	msg := fmt.Sprintf("holder:%s\nseats:%d\nexpires_at:%s", holder, seats, expiresAt)
	sig := ed25519.Sign(privKey, []byte(msg))
	sigHex := hex.EncodeToString(sig)

	// Write mock tormentnexus.lic
	licContent := fmt.Sprintf("holder: %s\nseats: %d\nexpires_at: %s\nsignature: %s\n", holder, seats, expiresAt, sigHex)
	licPath := filepath.Join(tempDir, "tormentnexus.lic")
	if err := os.WriteFile(licPath, []byte(licContent), 0644); err != nil {
		t.Fatalf("failed to write test license file: %v", err)
	}

	// Verify using VerifyLicense
	lic, err := VerifyLicense(tempDir)
	if err != nil {
		t.Fatalf("VerifyLicense failed: %v", err)
	}

	if lic.Holder != holder {
		t.Errorf("expected holder %q, got %q", holder, lic.Holder)
	}
	if lic.Seats != seats {
		t.Errorf("expected seats %d, got %d", seats, lic.Seats)
	}
}
