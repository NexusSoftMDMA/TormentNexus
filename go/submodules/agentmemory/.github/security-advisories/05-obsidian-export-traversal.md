# GHSA Draft: Arbitrary filesystem write via Obsidian export in agentmemory

**Severity:** Medium · **CVSS 3.1:** 6.5 (`AV:N/AC:L/PR:L/UI:N/S:U/C:N/I:H/A:L`)
**CWE:** [CWE-22 — Improper Limitation of a Pathname to a Restricted Directory ('Path Traversal')](https://cwe.mitre.org/data/definitions/22.html), [CWE-73 — External Control of File Name or Path](https://cwe.mitre.org/data/definitions/73.html)
**Affected versions:** `< 0.8.2`
**Patched version:** `0.8.2`

## Summary

The `POST /agentmemory/obsidian/export` endpoint accepted a `vaultDir` parameter and passed it directly to `mkdir` and `writeFile` calls without any containment check. A caller could set `vaultDir` to any absolute path on the filesystem and agentmemory would create directories and write Markdown files there with the permissions of the process running the server.

```bash
# Example exploit payload (affected versions only)
curl -X POST http://localhost:3111/agentmemory/obsidian/export \
  -H "Content-Type: application/json" \
  -d '{"vaultDir": "/etc/cron.d"}'
```

The content written would be agentmemory's exported memories in Markdown format, but an attacker could craft specific memory content beforehand to plant arbitrary files.

## Impact

When chained with advisory #03 (default `0.0.0.0` binding) or advisory #04 (unauthenticated mesh), an attacker on the local network could write arbitrary files to any filesystem location the agentmemory process had write access to.

Possible exploitation paths:
- Write to `~/.ssh/authorized_keys` — SSH key injection
- Write to `/etc/cron.d/*` — cron job injection (if running as root)
- Write to `~/.bashrc` or shell rc files — code execution on next shell
- Overwrite any file the process could write to

## Patches

Fixed in **0.8.2**:

- New `AGENTMEMORY_EXPORT_ROOT` environment variable (default: `~/.agentmemory`)
- `vaultDir` now goes through `resolveVaultDir()` in `src/functions/obsidian-export.ts`:
  - Resolves the path with `path.resolve`
  - Checks `resolved === root || resolved.startsWith(root + path.sep)`
  - Returns `null` if the check fails, and the endpoint returns `{ success: false, error: "vaultDir must be inside AGENTMEMORY_EXPORT_ROOT" }`
- Default export is confined to `~/.agentmemory/vault`
- Tests added in `test/obsidian-export.test.ts` for both the custom-but-valid case and the rejection case

## Known limitations

`resolveVaultDir()` performs lexical containment only — it does not call `fs.realpathSync` / `fs.lstatSync`. A pre-existing symlink under `AGENTMEMORY_EXPORT_ROOT` that points outside the root can still be written through. Users who allow untrusted processes to create files inside `AGENTMEMORY_EXPORT_ROOT` should additionally run agentmemory inside a sandbox that forbids symlink creation, or file a follow-up issue requesting symlink-aware containment.

## Workarounds

Users on affected versions should:
1. **Disable the Obsidian export endpoint** by setting `OBSIDIAN_AUTO_EXPORT=false` (and avoid calling `/agentmemory/obsidian/export` manually)
2. Set `AGENTMEMORY_SECRET` so the endpoint requires bearer auth
3. Upgrade to 0.8.2

## References

- Fix PR: [#108](https://github.com/rohitg00/agentmemory/pull/108)
- Commit: [`cbaaf4f`](https://github.com/rohitg00/agentmemory/commit/cbaaf4f)

## Credit

@eng-pf
