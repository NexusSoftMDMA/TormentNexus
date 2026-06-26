package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

const encryptedPrefix = "enc:"

// ConfigEncryptor handles encryption/decryption of server config JSON.
// If no key is configured, configs are stored as plaintext.
type ConfigEncryptor struct {
	gcm cipher.AEAD
}

// NewConfigEncryptor creates an encryptor from a passphrase.
// An empty key disables encryption (passthrough mode) with a warning.
func NewConfigEncryptor(key string) *ConfigEncryptor {
	if key == "" {
		slog.Warn("encryption key not set - credentials will be stored in plaintext. Set ARC_RELAY_ENCRYPTION_KEY for production use.")
		return &ConfigEncryptor{}
	}
	// Derive a 32-byte key from the passphrase using SHA-256
	hash := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		// SHA-256 always produces a valid AES-256 key size
		panic("unreachable: " + err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic("unreachable: " + err.Error())
	}
	return &ConfigEncryptor{gcm: gcm}
}

// Encrypt encrypts plaintext config JSON. Returns the original bytes if no key is set.
func (e *ConfigEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	if e.gcm == nil {
		return plaintext, nil
	}
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}
	ciphertext := e.gcm.Seal(nonce, nonce, plaintext, nil)
	encoded := encryptedPrefix + base64.StdEncoding.EncodeToString(ciphertext)
	return []byte(encoded), nil
}

// Decrypt decrypts config data. Handles both encrypted and legacy plaintext configs.
func (e *ConfigEncryptor) Decrypt(data []byte) ([]byte, error) {
	s := string(data)
	if !strings.HasPrefix(s, encryptedPrefix) {
		// Legacy plaintext config — return as-is
		return data, nil
	}
	if e.gcm == nil {
		return nil, fmt.Errorf("config is encrypted but no encryption key is configured")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(s, encryptedPrefix))
	if err != nil {
		return nil, fmt.Errorf("decoding encrypted config: %w", err)
	}
	nonceSize := e.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("encrypted config too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypting config: %w", err)
	}
	return plaintext, nil
}
