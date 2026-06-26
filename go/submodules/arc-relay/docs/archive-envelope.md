# Archive Envelope Schema

This is the wire contract between the Arc Relay archive middleware
(sender) and any receiver that accepts Arc Relay archive payloads -
most commonly the Comma Compliance ingest endpoint, but any operator
may host their own receiver that implements this schema.

Arc Relay is open source. The relay only stores the **recipient's
public key**; the matching private key lives on the receiver side and
never leaves it. Envelope encryption is an additional layer on top of
TLS so that a proxy, load balancer, or on-disk log in front of the
ingest endpoint cannot read payload bodies.

## Algorithm

NaCl Box as defined by libsodium:

- Key exchange: **Curve25519** (X25519)
- Authenticated encryption: **XSalsa20-Poly1305**
- Go reference: `golang.org/x/crypto/nacl/box`
- libsodium / Ruby / Python: `crypto_box_easy` / `RbNaCl::Box` / `nacl.public.Box`

The sender generates a **fresh ephemeral keypair for every payload**,
uses its private half to seal the box, and discards the private half
when the seal completes. The ephemeral public half travels in the
envelope so the receiver can complete the shared secret.

Because the sender private key is ephemeral, there is no long-lived
secret material on the Arc Relay side - just the recipient's public
key in the stored archive config.

## Wire format

All base64 fields use **standard** base64 (not URL-safe). The receiver
should reject any envelope whose fields do not decode cleanly.

```json
{
  "version":         "nacl-box-v1",
  "kid":             "<base64, 8 bytes>",
  "nonce":           "<base64, 24 bytes>",
  "ciphertext":      "<base64, opaque>",
  "sourcePublicKey": "<base64, 32 bytes>"
}
```

### Field semantics

| Field             | Required | Notes                                                                                      |
| ----------------- | -------- | ------------------------------------------------------------------------------------------ |
| `version`         | yes      | Must be `"nacl-box-v1"` for this schema. Receivers dispatch on this value (not ciphertext) |
| `kid`             | yes      | Fingerprint of the recipient public key. See below.                                        |
| `nonce`           | yes      | 24 random bytes, base64. Unique per payload; never reused.                                 |
| `ciphertext`      | yes      | NaCl Box sealed ciphertext of the plaintext archive payload.                               |
| `sourcePublicKey` | yes      | The ephemeral sender public key for this payload (32 bytes, base64).                       |

### Key ID (`kid`)

The `kid` is a stable fingerprint of the recipient's Curve25519 public
key, used by the receiver to pick the right private key during
rotation:

```
kid = base64( blake2b-256(recipient_pub)[:8] )
```

- Hash: **blake2b-256** over the raw 32 bytes of the public key
- Take the first 8 bytes of the digest
- Base64 encode with standard padding

The receiver and the Arc Relay side must compute `kid` identically,
because the relay stamps each envelope with the `kid` it computed from
its configured public key and the receiver looks up its matching
private key by that same `kid`.

### Plaintext (inside the ciphertext)

The plaintext sealed inside the ciphertext is the legacy archive
payload JSON that Arc Relay would otherwise POST directly when no
recipient key is configured. Schema:

```json
{
  "version":   "v1",
  "source":    "arc_relay",
  "phase":     "request" | "response" | "exchange" | "test",
  "timestamp": "2026-04-08T15:04:05Z",
  "meta": {
    "server_id":   "...",
    "server_name": "...",
    "user_id":     "...",
    "client_ip":   "...",
    "method":      "tools/call",
    "tool_name":   "...",
    "request_id":  "..."
  },
  "request":  { /* raw MCP request JSON-RPC object */ },
  "response": { /* raw MCP response JSON-RPC object */ }
}
```

Phase `"test"` is a synthetic payload emitted by the `/api/archive/test`
handler to exercise the full delivery path. Receivers should treat it
like any other payload (auth + decrypt + parse) and return 200, then
no-op downstream processing. Do not implement a special-case code path
for the test phase - that defeats the purpose of the drill.

## Legacy plaintext path

Existing receivers that do not yet implement envelope decryption must
continue to accept **plaintext** archive payloads. A plaintext payload
is the legacy schema above (top-level `version: "v1"`), POSTed
directly as the request body.

Receivers dispatch on the top-level `version` field:

- `"nacl-box-v1"` -> envelope path, decrypt with the private key
  matching `kid`
- `"v1"` (or missing) -> legacy plaintext path

**Do not** dispatch on the presence of `ciphertext`. Explicit version
bumps are how new envelope schemas will coexist without ambiguity.

## Rotation

Rotating a tenant's recipient key on the receiver side is a
multi-step, non-atomic operation because envelopes in the Arc Relay
delivery queue are **sealed at enqueue time**, not at delivery time.
Held rows cannot be re-sealed to a new key because the relay discards
plaintext as soon as the envelope is built.

Rotation contract:

1. The receiver generates a new keypair and marks it current. The old
   key enters a **grace period** during which its private key is kept
   and still used for inbound decryption.
2. Operators re-run the handoff on each Arc Relay instance that
   archives to this tenant. The handoff delivers the new public key
   and `kid`. Arc Relay seals all new envelopes to the new key.
3. During the grace period, the receiver tries the current private
   key first for each envelope. On a `kid` mismatch, it falls back to
   the private key whose `kid` matches.
4. After the grace period expires, the old private key is destroyed.
   Any envelopes still in an Arc Relay queue sealed to the old key
   will fail decrypt and should be cleared from the queue via
   "Clear Backlog" on the server detail page.

A grace period of 7 days is a reasonable default. Shorter is
operationally painful; longer leaves stale keys around.

## Security properties

- **Confidentiality from the transport layer down:** TLS terminators,
  reverse proxies, WAFs, and on-disk access logs in front of the
  receiver see only the envelope, not the sealed plaintext.
- **Forward secrecy within a single payload:** the sender private key
  is ephemeral, so a future compromise of the receiver's private key
  cannot decrypt past captured traffic - except insofar as the
  receiver's private key was the symmetric counterpart for the
  ephemeral exchange; NaCl Box is not forward-secret the way a full
  ECDHE handshake is. Treat this as a defense in depth feature, not
  as a substitute for rotation.
- **Authenticity:** NaCl Box is authenticated. A tampered envelope
  fails `box.Open`.
- **What this does NOT do:** envelope encryption does not
  authenticate the *sender*. An attacker who has stolen a valid
  bearer token can POST any sealed envelope they want - NaCl Box only
  binds the ephemeral sender pubkey to the ciphertext, and the
  ephemeral key is unverifiable. Receiver-side bearer token auth is
  the identity boundary.

## Non-goals

- No long-lived sender signing key. If compliance later needs per-
  instance provenance on envelopes, add a signature layer over the
  envelope, do not repurpose `sourcePublicKey`.
- No envelope key agreement or session keys. Every payload is an
  independent box with a fresh ephemeral sender key.
- No at-rest re-sealing of queued rows. If a tenant rotates keys
  while the Arc Relay queue has held rows, the operator is expected
  to either let the grace period cover delivery or clear the backlog.

## Reference implementations

- **Sender (Go):** `internal/middleware/archive_encrypt.go` -
  `encryptPayload`, `sealArchivePayload`, `ComputeKeyID`.
- **Receiver:** any HTTP endpoint that validates the envelope schema
  above and decrypts with the matching NaCl Box private key. See also
  [docs/archive-handoff.md](archive-handoff.md) for the protocol used
  to provision the public key into Arc Relay in the first place.
