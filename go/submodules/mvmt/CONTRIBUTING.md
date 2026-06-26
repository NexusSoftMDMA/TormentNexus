# Contributing

Thanks for taking the time to contribute to mvmt.

> **Using a coding agent?** Read [CLAUDE.md](CLAUDE.md) before making any changes. It contains rules for diffs, verification, and PR format that apply to all AI coding tools.

mvmt is a local personal data plane. The project is intentionally narrow: users choose exactly which local folders and native connectors are exposed, and mvmt gates access through MCP-compatible transports.

## Product Boundary

In scope:

- Explicit user-selected filesystem folder access.
- Native local-data connectors, starting with Obsidian.
- Security controls around local and tunneled access.
- Clear diagnostics, docs, and tests.

Out of scope for now:

- Connector marketplaces or package registries.
- Bundled SaaS/API connectors such as GitHub, Stripe, Linear, or similar services.
- Importing existing MCP configs from Claude Desktop, Claude Code, Cursor, or other tools.
- Production remote relay behavior before the security model is designed.

## Development Setup

```bash
nvm use
npm install
npm run verify
```

mvmt targets Node 20+. The repo includes `.nvmrc` with Node 20 for contributors who use nvm.

For local CLI development:

```bash
npm run dev -- --help
npm run dev -- init
npm run dev -- start
```

## Before Opening A Pull Request

Run:

```bash
npm run verify
```

GitHub Actions runs `npm ci` and `npm run verify` on push and pull request for Node 20.x and 22.x.

`npm run verify` is the default PR gate. It runs the TypeScript build, full tests, whitespace checks, package dry-run, and a local runtime smoke test that starts mvmt against a fixture vault, checks `/health`, shuts it down, and confirms the port is closed.

Useful variants:

```bash
npm run verify:quick    # skips the runtime smoke test
npm run verify:release  # includes npm audit --omit=dev
```

## Test Matrix

| Area | Type | Command | Notes |
| --- | --- | --- | --- |
| Config schema/loader | unit | `npm test` | Validates defaults, bad configs, tunnel config |
| Token handling | unit | `npm test` | Covers token read/show/rotate helpers |
| Origin/auth helpers | unit | `npm test` | Covers localhost and configured origins |
| Router | unit | `npm test` | Covers tool namespacing and dispatch |
| Obsidian connector | unit/integration fixture | `npm test` | Uses fixture vault data, not a real user vault |
| Proxy connector behavior | unit/mocked integration | `npm test` | Tests proxy behavior without requiring real third-party services |
| Doctor command | unit/mocked integration | `npm test` | Checks diagnostics behavior |
| Tunnel utilities | unit/process integration | `npm test` | Spawns short-lived local Node processes, not real tunnel providers |
| Runtime smoke | local integration | `npm run verify` | Starts mvmt on a temporary local port and verifies shutdown |
| TypeScript build | static | `npm run verify` | Required before PR |
| Package dry-run | packaging | `npm run verify` | Ensures npm package contents are sane |
| Dependency audit | network/security | `npm run verify:release` | Release check; not run in PR CI because it depends on registry availability |
| Coverage | coverage | `npm run test:coverage` | Optional before PR, useful for security-sensitive changes |
| Token + auth e2e | docker integration | `npm run e2e` | Builds the package, installs it into a fresh `node:20-alpine` container, exercises CLI + HTTP flows. Requires Docker. Excluded from `npm test`. |

## Platform Matrix

| Platform | Status | Notes |
| --- | --- | --- |
| macOS | primary | Main development platform |
| Linux | expected | Node 20+ support expected; needs regular CI coverage |
| Windows | partial | Path handling and chmod behavior need more coverage |

Connector tests should not require a real Obsidian vault, a real tunnel account, or production secrets. Use fixtures, mocks, or a temporary directory.

If your change touches security-sensitive behavior, add focused tests for the behavior and failure case. Security-sensitive areas include:

- Token generation or validation.
- OAuth/tunnel auth flow.
- Origin checks.
- Write gates.
- Filesystem path validation.
- Environment scrubbing.
- Audit logging.
- Connector process spawning.

## Pull Request Guidelines

- Keep PRs focused on one feature, bug fix, or documentation change.
- Explain the user-facing behavior change.
- Include test coverage proportional to the risk.
- Do not include real vault paths, API keys, bearer tokens, or personal local config.
- Prefer explicit scope and read-only defaults for any connector or file access change.

## Contributor License

No contributor license agreement is required right now. Contributions are accepted under the repository's MIT license. If the project starts accepting broad external contributions or changes its distribution model, this may be revisited before larger contributions are merged.

## Commit Style

Use short, direct commit messages:

```text
add token rotation command
fix obsidian path traversal check
document tunnel limitations
```

If this is done by a coding agent, prepend with the name of the agent:
```test
[codex] Refactor config.ts
```

## Security Changes

Security changes should be conservative. Use common sense and double check. When in doubt:

- Default to less access.
- Require explicit user consent for writes.
- Document the limitation rather than hiding it.
- Add tests for rejection paths, not just happy paths.

Please do not disclose vulnerabilities in public issues. See `SECURITY.md`.
