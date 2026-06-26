# GHSA Draft: Unauthenticated mesh sync in agentmemory

**Severity:** High · **CVSS 3.1:** 7.4 (`AV:N/AC:H/PR:N/UI:N/S:U/C:H/I:H/A:N`)
**CWE:** [CWE-306 — Missing Authentication for Critical Function](https://cwe.mitre.org/data/definitions/306.html), [CWE-862 — Missing Authorization](https://cwe.mitre.org/data/definitions/862.html)
**Affected versions:** `< 0.8.2`
**Patched version:** `0.8.2`

## Summary

agentmemory's mesh federation feature (P2P sync between instances) accepted push/pull requests on its `/agentmemory/mesh/*` endpoints without requiring authentication. The mesh sync function also did not send any `Authorization` header when calling peer instances, meaning the federation protocol was entirely unauthenticated.

## Impact

Any attacker who could reach a mesh-enabled agentmemory instance could:

1. **Push fake memories** via `POST /agentmemory/mesh/receive` — inject attacker-controlled observations, actions, semantic memories, and relations into the target's memory store. This poisons future retrievals and could be used to manipulate what the target's AI agent sees.
2. **Pull the entire memory store** via `GET /agentmemory/mesh/export` — download all memories, actions, and graph data marked as mesh-shareable.
3. **Chain with advisory #03** — combined with the default `0.0.0.0` binding, mesh endpoints were reachable from any device on the local network without any authentication.

Mesh is opt-in (requires an explicit peer registration), so this affected only users who had enabled federation. But those users had no authentication at all.

## Patches

Fixed in **0.8.2**:

- All 5 mesh REST endpoints (`mesh-register`, `mesh-list`, `mesh-sync`, `mesh-receive`, `mesh-export`) now return 503 with `"mesh requires AGENTMEMORY_SECRET"` if the secret is not configured
- The `mem::mesh-sync` function now accepts a `meshAuthToken` parameter and **refuses to sync at all** if the token is missing
- Outgoing push/pull requests include `Authorization: Bearer <secret>` headers
- Server-side, all mesh endpoints check bearer auth via the existing `checkAuth` helper

## Workarounds

Users on affected versions who have mesh federation enabled should:
1. Set `AGENTMEMORY_SECRET` to a strong random value on **both** peers
2. Restart the server
3. Upgrade to 0.8.2 at the earliest opportunity

Users who have **not** enabled mesh federation are not affected by this specific issue, but should still upgrade for the other 5 fixes.

## References

- Fix PR: [#108](https://github.com/rohitg00/agentmemory/pull/108)
- Commit: [`cbaaf4f`](https://github.com/rohitg00/agentmemory/commit/cbaaf4f)

## Credit

@eng-pf
