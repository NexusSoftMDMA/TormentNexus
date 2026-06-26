package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/box"
)

// EnvelopeVersion identifies the envelope schema. Compliance receivers
// dispatch on this value rather than on the presence of a ciphertext
// field so new envelope schemas can coexist with legacy plaintext posts.
const EnvelopeVersion = "nacl-box-v1"

// naclEnvelope is the encrypted payload envelope sent to the archive webhook.
// Wire format (all base64 fields use standard encoding):
//
//	{
//	  "version":         "nacl-box-v1",
//	  "kid":             "<base64 8-byte fingerprint>",
//	  "nonce":           "<base64 24 bytes>",
//	  "ciphertext":      "<base64 opaque>",
//	  "sourcePublicKey": "<base64 32 bytes ephemeral>"
//	}
type naclEnvelope struct {
	Version         string `json:"version"`
	KeyID           string `json:"kid,omitempty"`
	Nonce           string `json:"nonce"`
	Ciphertext      string `json:"ciphertext"`
	SourcePublicKey string `json:"sourcePublicKey"`
}

// ComputeKeyID derives a stable fingerprint for a recipient public key.
// Compliance must compute this the same way so the ingest side can look
// up the matching private key during rotation (see docs/archive-envelope.md).
// Algorithm: first 8 bytes of blake2b-256(pubkey), base64-encoded.
func ComputeKeyID(pubKey [32]byte) string {
	sum := blake2b.Sum256(pubKey[:])
	return base64.StdEncoding.EncodeToString(sum[:8])
}

// encryptPayload encrypts a JSON payload using NaCl Box (X25519 + XSalsa20-Poly1305)
// with an ephemeral sender keypair. Returns the JSON-encoded envelope. The
// sender private key is generated fresh per call and is discarded when this
// function returns - nothing persistent lives on the relay side.
func encryptPayload(payload []byte, recipientPubKey [32]byte) ([]byte, error) {
	senderPub, senderPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating ephemeral keypair: %w", err)
	}

	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	ciphertext := box.Seal(nil, payload, &nonce, &recipientPubKey, senderPriv)

	envelope := naclEnvelope{
		Version:         EnvelopeVersion,
		KeyID:           ComputeKeyID(recipientPubKey),
		Nonce:           base64.StdEncoding.EncodeToString(nonce[:]),
		Ciphertext:      base64.StdEncoding.EncodeToString(ciphertext),
		SourcePublicKey: base64.StdEncoding.EncodeToString(senderPub[:]),
	}

	return json.Marshal(envelope)
}

// sealArchivePayload returns the sealed envelope when recipientKey is
// non-nil, otherwise returns the payload unchanged. This is the single
// sealing entry point shared by both the delivery-queue enqueue path
// (Archive.enqueue) and the synchronous test-delivery path
// (ArchiveDispatcher.SendTest) so the two stay in lockstep.
func sealArchivePayload(payload []byte, recipientKey *[32]byte) ([]byte, error) {
	if recipientKey == nil {
		return payload, nil
	}
	return encryptPayload(payload, *recipientKey)
}

// DecodeRecipientKey decodes a base64-encoded Curve25519 public key.
// Exported so other packages (e.g. web handlers rendering a key
// fingerprint) can reuse the same length-checked decode used on the
// hot path here.
func DecodeRecipientKey(b64Key string) ([32]byte, error) {
	var key [32]byte
	raw, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		return key, fmt.Errorf("invalid base64: %w", err)
	}
	if len(raw) != 32 {
		return key, fmt.Errorf("expected 32 bytes, got %d", len(raw))
	}
	copy(key[:], raw)
	return key, nil
}
