package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"

	"golang.org/x/crypto/nacl/box"
)

func TestEncryptPayloadRoundTrip(t *testing.T) {
	// Generate recipient keypair
	recipientPub, recipientPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generating recipient keypair: %v", err)
	}

	payload := []byte(`{"version":"v1","source":"arc_relay","phase":"test"}`)

	// Encrypt
	envelopeJSON, err := encryptPayload(payload, *recipientPub)
	if err != nil {
		t.Fatalf("encryptPayload: %v", err)
	}

	// Parse envelope
	var envelope naclEnvelope
	if err := json.Unmarshal(envelopeJSON, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}

	// Verify all fields present
	if envelope.Version != EnvelopeVersion {
		t.Errorf("version = %q, want %q", envelope.Version, EnvelopeVersion)
	}
	if envelope.KeyID == "" {
		t.Error("kid is empty")
	}
	if envelope.Nonce == "" {
		t.Error("nonce is empty")
	}
	if envelope.Ciphertext == "" {
		t.Error("ciphertext is empty")
	}
	if envelope.SourcePublicKey == "" {
		t.Error("sourcePublicKey is empty")
	}
	// kid must match the computed fingerprint of the recipient pubkey
	expectedKID := ComputeKeyID(*recipientPub)
	if envelope.KeyID != expectedKID {
		t.Errorf("kid = %q, want %q (derived from recipient pubkey)", envelope.KeyID, expectedKID)
	}

	// Decode and decrypt
	nonce, err := base64.StdEncoding.DecodeString(envelope.Nonce)
	if err != nil {
		t.Fatalf("decode nonce: %v", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(envelope.Ciphertext)
	if err != nil {
		t.Fatalf("decode ciphertext: %v", err)
	}
	senderPub, err := base64.StdEncoding.DecodeString(envelope.SourcePublicKey)
	if err != nil {
		t.Fatalf("decode sender pub: %v", err)
	}

	var nonceArr [24]byte
	copy(nonceArr[:], nonce)
	var senderPubArr [32]byte
	copy(senderPubArr[:], senderPub)

	decrypted, ok := box.Open(nil, ciphertext, &nonceArr, &senderPubArr, recipientPriv)
	if !ok {
		t.Fatal("box.Open failed - decryption error")
	}

	if string(decrypted) != string(payload) {
		t.Errorf("decrypted = %q, want %q", string(decrypted), string(payload))
	}
}

func TestDecodeRecipientKey(t *testing.T) {
	// Valid 32-byte key
	pub, _, _ := box.GenerateKey(rand.Reader)
	b64 := base64.StdEncoding.EncodeToString(pub[:])

	key, err := DecodeRecipientKey(b64)
	if err != nil {
		t.Fatalf("DecodeRecipientKey: %v", err)
	}
	if key != *pub {
		t.Error("decoded key does not match original")
	}

	// Invalid base64
	_, err = DecodeRecipientKey("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}

	// Wrong length
	short := base64.StdEncoding.EncodeToString([]byte("tooshort"))
	_, err = DecodeRecipientKey(short)
	if err == nil {
		t.Error("expected error for wrong length key")
	}
}

func TestEncryptPayloadDifferentNonces(t *testing.T) {
	pub, _, _ := box.GenerateKey(rand.Reader)
	payload := []byte(`{"test": true}`)

	env1, _ := encryptPayload(payload, *pub)
	env2, _ := encryptPayload(payload, *pub)

	var e1, e2 naclEnvelope
	if err := json.Unmarshal(env1, &e1); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(env2, &e2); err != nil {
		t.Fatal(err)
	}

	if e1.Nonce == e2.Nonce {
		t.Error("two encryptions produced the same nonce - nonces must be unique")
	}
	if e1.SourcePublicKey == e2.SourcePublicKey {
		t.Error("two encryptions used the same ephemeral key - keys must be unique per message")
	}
	// kid is derived from the recipient pubkey so it must NOT change
	// between encryptions using the same recipient. Compliance routes
	// on kid during rotation, so instability would break delivery.
	if e1.KeyID != e2.KeyID {
		t.Errorf("kid differs across encryptions: %q vs %q (must be stable for same recipient)", e1.KeyID, e2.KeyID)
	}
}

func TestComputeKeyIDStability(t *testing.T) {
	// Same pubkey must produce the same kid every time.
	pub, _, _ := box.GenerateKey(rand.Reader)
	kid1 := ComputeKeyID(*pub)
	kid2 := ComputeKeyID(*pub)
	if kid1 != kid2 {
		t.Errorf("kid not stable: %q vs %q", kid1, kid2)
	}
	// Different pubkeys must produce different kids (with overwhelming probability).
	other, _, _ := box.GenerateKey(rand.Reader)
	if ComputeKeyID(*other) == kid1 {
		t.Error("different pubkeys collided on kid")
	}
	// kid must decode to 8 bytes.
	raw, err := base64.StdEncoding.DecodeString(kid1)
	if err != nil {
		t.Fatalf("kid is not valid base64: %v", err)
	}
	if len(raw) != 8 {
		t.Errorf("kid decodes to %d bytes, want 8", len(raw))
	}
}

func TestSealArchivePayloadPassthrough(t *testing.T) {
	// Nil recipient key means encryption is not configured; sealArchivePayload
	// must return the payload verbatim so enqueue still works for plaintext
	// tenants on the legacy path.
	payload := []byte(`{"hello":"world"}`)
	sealed, err := sealArchivePayload(payload, nil)
	if err != nil {
		t.Fatalf("sealArchivePayload: %v", err)
	}
	if string(sealed) != string(payload) {
		t.Errorf("passthrough changed payload: got %q, want %q", sealed, payload)
	}
}

func TestSealArchivePayloadSeals(t *testing.T) {
	pub, _, _ := box.GenerateKey(rand.Reader)
	payload := []byte(`{"hello":"world"}`)
	sealed, err := sealArchivePayload(payload, pub)
	if err != nil {
		t.Fatalf("sealArchivePayload: %v", err)
	}
	if string(sealed) == string(payload) {
		t.Error("sealArchivePayload returned plaintext when a key was provided")
	}
	var env naclEnvelope
	if err := json.Unmarshal(sealed, &env); err != nil {
		t.Fatalf("sealed output is not a valid envelope: %v", err)
	}
	if env.Version != EnvelopeVersion {
		t.Errorf("version = %q, want %q", env.Version, EnvelopeVersion)
	}
	if env.Ciphertext == "" {
		t.Error("ciphertext is empty")
	}
}
