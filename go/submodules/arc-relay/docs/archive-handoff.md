# Archive Handoff Protocol

This describes the stateful handoff flow used to provision a Comma
Compliance archive endpoint into an Arc Relay instance. It is the
protocol a receiver (e.g., `commacompliance.ai`) must implement to
interoperate with Arc Relay's "Set up the Comma Compliance Archive"
flow on the server detail page.

The handoff solves three problems at once:

1. Point the archive middleware at a URL the receiver controls.
2. Provision a bearer token that the receiver will accept.
3. Provision the NaCl recipient public key + key ID used for
   [envelope encryption](archive-envelope.md).

All three values arrive in a single user-initiated round trip so
operators do not have to paste anything by hand.

## Threat model

The handoff is interactive and browser-driven, so the main threats
are:

- **Crafted fragments:** an attacker gets an authenticated admin to
  visit a URL like `#mw-archive?archive_url=evil&archive_token=pwn&...`
  and silently reconfigures their archive destination to attacker-
  controlled infrastructure.
- **Open redirect on the receiver side:** the receiver accepts a
  `return_to` parameter from the relay and might bounce a user to an
  attacker URL with credentials in the fragment.
- **Fragment replay:** an attacker replays a legitimate handoff
  fragment at a different time or on a different user's session.

The protocol defends against all three with a **one-time nonce bound
to the initiating admin's session**. The relay never applies config
from a raw fragment; it requires a server-side consume step.

## Flow

```
ADMIN                   ARC RELAY                    COMPLIANCE RECEIVER
  |                        |                                  |
  |--click Set up--------->|                                  |
  |                        |--POST handoff/begin------------->|
  |                        |<-{state: NONCE, expires_in}------|
  |<--window.open(         |                                  |
  |   compliance URL?      |                                  |
  |   return_to=RELAY&     |                                  |
  |   state=NONCE)         |                                  |
  |                        |                                  |
  |----------------------- new tab -----------------------    |
  | sign up / configure archive on compliance ------------>   |
  |                                                            |
  |<--redirect to          |                                  |
  |   RELAY#mw-archive?    |                                  |
  |   state=NONCE&         |                                  |
  |   archive_url=...&     |                                  |
  |   archive_token=...&   |                                  |
  |   nacl_recipient_key=  |                                  |
  |   ...&nacl_key_id=...  |                                  |
  |                        |                                  |
  |--JS reads fragment     |                                  |
  |--POST handoff/complete |                                  |
  |  {state, ...values}--->|                                  |
  |                        |--validate nonce, save config     |
  |                        |<--{status: ok, config: {...}}----|
  |                        |--rewrite queue, run test---------|
```

### Step 1: Begin

Arc Relay endpoint: `POST /api/archive/handoff/begin`

Admin-only. No request body needed. Response:

```json
{
  "state":      "<random 32-byte base64url nonce>",
  "expires_in": 600
}
```

The relay stores `(state -> user_id, expires_at)` in an in-memory
store for `archiveHandoffTTL` (10 minutes). The same nonce cannot be
issued twice and cannot be used by a different session.

### Step 2: Open compliance popup

The relay's JavaScript opens a new tab:

```
https://commacompliance.ai/compliance-archive
  ?return_to=<url-encoded relay origin + path>
  &state=<nonce from step 1>
```

The receiver is expected to:

- Validate `return_to` against its own allowlist or sanity check (not
  open-redirect). A minimal allowlist: "must be https (or localhost),
  must be a URL the current tenant registered as a relay endpoint".
- Echo `state` verbatim back in the return redirect. The relay treats
  `state` as opaque.
- Complete whatever signup / configuration flow it needs with the
  admin.

### Step 3: Return redirect

When configuration is complete, the receiver redirects the browser to:

```
<return_to>#mw-archive?
  state=<nonce>&
  archive_url=<url-encoded ingest URL>&
  archive_token=<url-encoded bearer token>&
  nacl_recipient_key=<base64 32-byte Curve25519 public key>&
  nacl_key_id=<base64 8-byte blake2b-256 fingerprint>
```

All five values live in the **hash fragment**, not the query string.
Hash fragments are client-only - they never reach the server, never
appear in `Referer` headers, and are not written to HTTP access logs.

Base64 encoding is **standard** (not URL-safe). The relay's decoder
uses `base64.StdEncoding`.

`nacl_key_id` is advisory: the relay recomputes it from the pubkey
using the algorithm in [archive-envelope.md](archive-envelope.md) so
a fraudulent or stale key ID cannot poison rotation routing on the
receiver.

### Step 4: Complete

The relay's JS reads the fragment and POSTs to
`POST /api/archive/handoff/complete`:

```json
{
  "state":              "<nonce>",
  "archive_url":        "https://...",
  "archive_token":      "...",
  "nacl_recipient_key": "<base64>",
  "nacl_key_id":        "<base64>"
}
```

Server-side the relay:

1. Validates the nonce. Unknown, expired, or cross-user nonces all
   return the same error (`"handoff expired or invalid, please retry
   setup"`) so attackers learn nothing about the store's contents.
2. The nonce is consumed on lookup regardless of outcome, so a failed
   attempt cannot be retried.
3. Merges the fragment values with any existing archive config so
   settings like `include` and `api_key_header` that the operator
   customized locally are preserved across re-runs.
4. Runs `ValidateArchiveConfig` on the merged config, rejecting bad
   URLs, unknown auth types, or malformed NaCl keys.
5. Normalizes and persists the config.
6. Rewrites queued/held archive rows onto the new URL and resets the
   circuit breaker so a previously-paused dispatcher resumes against
   the new destination.
7. Returns the saved config (minus secrets, plus the derived `kid`)
   for the UI to re-render the form.

### Step 5: Post-complete test

Immediately after a successful complete, the relay UI POSTs to
`/api/archive/test`. When a recipient key is configured, the test
path seals the test payload with the same envelope code the real
delivery path uses. A bad key surfaces as a test failure at handoff
time, not at 3am on the first real request.

## Downgrade to plaintext

An admin can clear a previously saved NaCl key by:

- Clicking **Remove encryption** on the envelope indicator row in
  the archive config form. The hidden input is cleared, the form is
  re-saved immediately, and the downgrade is persisted in the same
  request.
- Re-running the handoff with a compliance receiver that does not
  include `nacl_recipient_key` in the return fragment. An empty key
  in the fragment explicitly clears any stored key.

Both paths are explicit to avoid sticky encryption state where the
UI claims "not encrypted" while the server still holds an old key.

## Non-goals

- **No signing of the fragment.** NaCl Box envelopes already
  authenticate the ciphertext. The fragment values arrive in the
  operator's own browser, initiated by their own click. A signature
  over the fragment would add ceremony without closing a real attack
  path, as long as the nonce protocol is in place.
- **No multi-destination provisioning in a single handoff.** Each
  handoff configures the single global archive middleware. To
  provision multiple destinations, run the handoff multiple times.

## Operator-visible failure modes

| Symptom                                                | Likely cause                                                               |
| ------------------------------------------------------ | -------------------------------------------------------------------------- |
| "handoff expired or invalid, please retry setup"       | Nonce was consumed, expired, or came from a different session              |
| "archive: invalid nacl_recipient_key: expected 32..." | Compliance returned a malformed or wrong-algorithm key                     |
| "archive: url must use https..."                       | Compliance returned a plain-http URL for a non-localhost host               |
| Test delivery fails after successful handoff save     | Bearer token is wrong, receiver rejects the request, or envelope kid does not match receiver's current private key |

## Reference implementation

- **Relay server:** `internal/web/archive_handoff.go` -
  `handleArchiveHandoffBegin`, `handleArchiveHandoffComplete`,
  `archiveHandoffStore`.
- **Relay UI:** `internal/web/templates/server_detail.html` -
  `openArchiveSetup` (step 2), hash parser IIFE at the bottom of the
  `<script>` block (step 4).
- **Receiver:** implemented out of tree by the archive recipient;
  the contract is the begin/complete endpoints documented above.
