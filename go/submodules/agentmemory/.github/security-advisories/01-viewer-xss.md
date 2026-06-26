# GHSA Draft: Stored XSS in agentmemory real-time viewer

**Severity:** Critical · **CVSS 3.1:** 9.6 (`AV:N/AC:L/PR:N/UI:R/S:C/C:H/I:H/A:L`)
**CWE:** [CWE-79 — Improper Neutralization of Input During Web Page Generation](https://cwe.mitre.org/data/definitions/79.html)
**Affected versions:** `< 0.8.2`
**Patched version:** `0.8.2`

## Summary

agentmemory's real-time viewer (default port 3113) rendered user-controlled data — tool outputs, file paths, memory titles, observation content — into HTML using inline `onclick=` event handlers. The viewer's Content Security Policy simultaneously allowed `script-src 'unsafe-inline'`, meaning injected JavaScript would execute in the reader's browser context.

## Impact

Any data captured by agentmemory hooks — which includes tool output from Claude Code, Cursor, or any other agent — becomes an XSS vector when the user opens the viewer. An attacker with the ability to influence any captured observation (e.g., by sending a crafted file contents to be read by an agent, or by planting a malicious commit message in a repository) could:

- Exfiltrate the entire memory store via authenticated requests from the browser
- Read `AGENTMEMORY_SECRET` if the viewer was configured with auth
- Make requests to arbitrary endpoints on behalf of the viewer user
- Modify the DOM to mislead the developer
- Pivot to other localhost services on the developer's machine

The viewer runs on localhost by default but is **reachable from the browser**, so standard same-origin protections don't help.

## Patches

Fixed in **0.8.2**:

- All inline `on*=` handlers removed from `src/viewer/index.html`
- Replaced with delegated `data-action` event handling
- CSP switched to a **per-response script nonce** (`script-src 'nonce-<random>'`)
- Added `script-src-attr 'none'` to block any inline handler attributes even if injected
- Viewer HTML now rendered through `src/viewer/document.ts` which generates a fresh nonce per request

## Workarounds

**None.** Users on affected versions should upgrade to 0.8.2 immediately. Do not open `http://localhost:3113` in a browser on affected versions if you suspect any of your captured observations may contain attacker-controlled content.

## References

- Fix PR: [#108](https://github.com/rohitg00/agentmemory/pull/108)
- Commit: [`cbaaf4f`](https://github.com/rohitg00/agentmemory/commit/cbaaf4f)
- Reporter: @eng-pf

## Credit

@eng-pf submitted PR #108 with fixes for this and 5 other vulnerabilities.
