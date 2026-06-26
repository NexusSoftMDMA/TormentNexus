# GHSA Draft: Remote shell script execution in agentmemory CLI startup

**Severity:** Critical · **CVSS 3.1:** 9.8 (`AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H`)
**CWE:** [CWE-494 — Download of Code Without Integrity Check](https://cwe.mitre.org/data/definitions/494.html), [CWE-829 — Inclusion of Functionality from Untrusted Control Sphere](https://cwe.mitre.org/data/definitions/829.html)
**Affected versions:** `< 0.8.2`
**Patched version:** `0.8.2`

## Summary

The agentmemory CLI (`npx @agentmemory/agentmemory`) auto-installed the iii-engine binary by piping a remote shell script into `sh`:

```ts
execSync("curl -fsSL https://install.iii.dev/iii/main/install.sh | sh")
```

This happened automatically on first run if `iii` was not found in `$PATH`. The script was fetched over HTTPS and executed with the permissions of the user running `npx agentmemory`. No checksum verification, no pinned version, no signature check.

## Impact

If `install.iii.dev` were ever compromised — via DNS hijack, domain takeover, expired certificate + MITM on an untrusted network, BGP attack, or any other supply chain attack — **every new agentmemory user would execute attacker-controlled shell code** as their own user.

This is the canonical "curl | sh" supply chain anti-pattern. It affected:
- Developers running `npx @agentmemory/agentmemory` for the first time
- CI/CD pipelines that installed agentmemory fresh
- Docker builds that installed agentmemory as part of an image

## Patches

Fixed in **0.8.2**:

- Removed `execSync` call entirely from `src/cli.ts`
- CLI now uses an existing local `iii` binary if present in `$PATH`
- Falls back to Docker Compose (`docker compose up -d`) if Docker is available
- Shows manual install instructions if neither iii nor Docker is found:
  - `cargo install iii-engine`
  - `docker pull iiidev/iii:latest`
  - Docs link: https://iii.dev/docs

## Workarounds

Users on affected versions should **install iii-engine manually** and run `agentmemory --no-engine` until upgraded:

```bash
cargo install iii-engine
npx @agentmemory/agentmemory@0.8.1 --no-engine
```

Then upgrade to 0.8.2 at the earliest opportunity.

## References

- Fix PR: [#108](https://github.com/rohitg00/agentmemory/pull/108)
- Commit: [`cbaaf4f`](https://github.com/rohitg00/agentmemory/commit/cbaaf4f)

## Credit

@eng-pf
