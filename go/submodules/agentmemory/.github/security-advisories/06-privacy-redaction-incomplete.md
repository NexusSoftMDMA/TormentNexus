# GHSA Draft: Incomplete secret redaction in agentmemory privacy filter

**Severity:** Medium · **CVSS 3.1:** 6.2 (`AV:L/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N`)
**CWE:** [CWE-532 — Insertion of Sensitive Information into Log File](https://cwe.mitre.org/data/definitions/532.html), [CWE-200 — Exposure of Sensitive Information to an Unauthorized Actor](https://cwe.mitre.org/data/definitions/200.html)
**Affected versions:** `< 0.8.2`
**Patched version:** `0.8.2`

## Summary

agentmemory's privacy filter (`src/functions/privacy.ts`) is supposed to strip API keys, secrets, and bearer tokens from captured observations before they are stored. The filter used regex patterns to detect common token formats. Three modern token formats were missing from the patterns:

1. **Bearer tokens** — `Authorization: Bearer <token>` headers were not matched, so any captured HTTP request or response that included an Authorization header flowed into the memory store verbatim.
2. **OpenAI project keys** — `sk-proj-*` (the dominant OpenAI API key format since mid-2024) was not matched. The existing `sk-[A-Za-z0-9]{20,}` pattern only caught the legacy format.
3. **GitHub fine-grained service/user tokens** — `ghs_*` and `ghu_*` were not matched. The existing `ghp_[A-Za-z0-9]{36}` pattern only caught personal access tokens.

## Impact

agentmemory's README explicitly claimed "Privacy first — API keys, secrets, and `<private>` tags are stripped before anything is stored." That claim was **false** for three common token formats.

Users relying on the privacy filter to protect their captured observations had a false sense of security. Tokens matching these three patterns would:

1. Be captured by `PostToolUse` hooks alongside the rest of the tool output
2. Pass through `stripPrivateData()` unmodified
3. Be LLM-compressed and stored in the memory KV
4. Be exposed to any attacker who could reach the `/agentmemory/export` or `/agentmemory/smart-search` endpoints
5. Be included in Obsidian exports, mesh syncs, and CLAUDE.md bridge writes

When chained with advisory #03 (default `0.0.0.0` binding), this meant network-adjacent attackers could retrieve captured Bearer tokens, OpenAI keys, and GitHub service tokens from the memory store.

## Patches

Fixed in **0.8.2**:

New regex patterns added to `SECRET_PATTERN_SOURCES` in `src/functions/privacy.ts`:

```ts
/Bearer\s+[A-Za-z0-9._\-+/=]{20,}/gi,
/sk-proj-[A-Za-z0-9\-_]{20,}/g,
/(?:sk|pk|rk|ak)-[A-Za-z0-9][A-Za-z0-9\-_]{19,}/g,
/gh[pus]_[A-Za-z0-9]{36,}/g,
```

Three new unit tests in `test/privacy.test.ts` verify each format is now stripped.

## Workarounds

Users on affected versions should:
1. Avoid having agents read files or API responses containing these token formats
2. Use the `<private>` tag around any block containing secrets — that filter was not affected
3. Set `AGENTMEMORY_SECRET` to restrict API access
4. Upgrade to 0.8.2

## References

- Fix PR: [#108](https://github.com/rohitg00/agentmemory/pull/108)
- Commit: [`cbaaf4f`](https://github.com/rohitg00/agentmemory/commit/cbaaf4f)

## Credit

@eng-pf
