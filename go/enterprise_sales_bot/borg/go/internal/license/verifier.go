package license

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Hardcoded public key for Ed25519 license verification
const licensePublicKeyHex = "9a9d5d9cc7acebbbf80adfe9005586c3f6496e82e7fa300920b831397c1cb763"

// License represents the decrypted and validated license properties
type License struct {
	Holder    string
	Seats     int
	ExpiresAt time.Time
}

// Simple YAML parser to avoid external dependencies
func parseLicenseYAML(data []byte) map[string]string {
	m := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			m[k] = v
		}
	}
	return m
}

// VerifyLicense loads and validates a tormentnexus.lic signed YAML license block.
func VerifyLicense(workspaceRoot string) (*License, error) {
	licPath := filepath.Join(workspaceRoot, "tormentnexus.lic")
	data, err := os.ReadFile(licPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read license file: %w", err)
	}

	fields := parseLicenseYAML(data)
	holder := fields["holder"]
	seatsStr := fields["seats"]
	expiresStr := fields["expires_at"]
	sigHex := fields["signature"]

	if holder == "" || seatsStr == "" || expiresStr == "" || sigHex == "" {
		return nil, errors.New("license file is missing required fields (holder, seats, expires_at, signature)")
	}

	seats, err := strconv.Atoi(seatsStr)
	if err != nil {
		return nil, fmt.Errorf("invalid seats field: %w", err)
	}

	// Verify cryptographic signature using public key
	pubKeyBytes, err := hex.DecodeString(licensePublicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	pubKey := ed25519.PublicKey(pubKeyBytes)

	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return nil, fmt.Errorf("invalid signature hex encoding: %w", err)
	}

	// Message layout: "holder:{holder}\nseats:{seats}\nexpires_at:{expires_at}"
	msg := fmt.Sprintf("holder:%s\nseats:%d\nexpires_at:%s", holder, seats, expiresStr)

	if !ed25519.Verify(pubKey, []byte(msg), sigBytes) {
		return nil, errors.New("cryptographic signature verification failed")
	}

	// Validate expiration time
	expiresAt, err := time.Parse(time.RFC3339, expiresStr)
	if err != nil {
		return nil, fmt.Errorf("invalid expires_at field (must be RFC3339 format): %w", err)
	}

	if time.Now().After(expiresAt) {
		return nil, fmt.Errorf("license has expired on %s", expiresAt.Format(time.RFC822))
	}

	return &License{
		Holder:    holder,
		Seats:     seats,
		ExpiresAt: expiresAt,
	}, nil
}
