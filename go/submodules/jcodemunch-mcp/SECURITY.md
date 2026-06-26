# Security Controls

jcodemunch-mcp indexes source code from local folders and GitHub repositories. This document describes the security controls that protect against common risks when handling arbitrary codebases.

---

## Path Traversal Prevention

All user-supplied paths are validated before any file is read or written.

* **`validate_path(root, target)`** resolves both paths to absolute form and verifies the target is a descendant of `root` using `os.path.commonpath()`.
* Applied during file discovery and again before each file read (defense in depth).
* Paths such as `../../etc/passwd` or absolute paths outside the repository root are rejected.

---

## Symlink Escape Protection

Symlinks can be used to escape the repository root and read arbitrary files.

* **Default:** `follow_symlinks=False` — symlinks are skipped during file discovery.
* When symlinks are followed (`follow_symlinks=True`), each symlink target is resolved and validated against the repository root. Escaping symlinks are skipped with a warning.
* **`is_symlink_escape(root, path)`** checks whether a symlink resolves outside the root.
* On Windows, environments without symlink support automatically skip symlink traversal.

---

## Default Ignore Policy

Files are filtered through multiple layers:

1. **SKIP_PATTERNS** — directories and files always excluded (e.g., `node_modules/`, `vendor/`, `.git/`, `build/`, `dist/`, generated files, lock files).
2. **`.gitignore`** — respected by default for both local folders and GitHub repositories (via the `pathspec` library).
3. **`extra_ignore_patterns`** — user-configurable additional gitignore-style patterns passed to indexing tools.

---

## Secret Exclusion

Credential files are excluded during indexing by a structured classifier
(`secret_classifier.classify_secret_file`, behind the `security.is_secret_file`
boolean). It decides from the **filename and directory shape only** — it never
reads file contents. Detection is organized into groups, in precedence order:

* **Exact credential names:** `.env`, `.env.*`, `.htpasswd`, `.netrc`, `.npmrc`,
  `.pypirc`, `credentials.json`, `service-account*.json`, `client_secret*.json`,
  `application_default_credentials.json`, `*-firebase-adminsdk-*.json`,
  `token.json`, `*.token`, `*.credentials`, `*.secrets`, `*.agekey`, and SSH
  private keys (`id_rsa*`, `id_ed25519*`, `id_dsa*`, `id_ecdsa*`).
* **Credential extensions:** `*.pem`, `*.key`, `*.p8`, `*.p12`, `*.pfx`, `*.ppk`,
  `*.jks`, `*.keystore`.
* **Path-specific credentials:** `~/.aws/credentials`, `~/.kube/config`,
  `~/.docker/config.json`, `~/.config/gcloud/application_default_credentials.json`,
  `~/.azure/accessTokens.json`, `~/.cargo/credentials[.toml]`, `~/.gem/credentials`,
  `composer/auth.json` — credential stores that only a path (not a basename) names.
* **Key-material directories:** a private-key/keystore file under `keys/`,
  `certs/`, `ssl/`, `tls/`, `pki/`, … (private-key extensions only — **public**
  certs `*.crt`/`*.cer`/`*.der` are not excluded just for living here).
* **Secret-store data:** data/config files (`*.yaml`, `*.json`, `*.tfvars`,
  `*.tfstate[.backup]`, …) under a whole-segment `secret`/`secrets`/`credential`/
  `credentials`/`creds`/`vault` directory.
* **Broad `secret` basename:** a `secret`/`secrets` token at a **word boundary**
  (so `prod-secrets.yaml` is excluded but `secretariat.csv` is not), excluding
  source-code and documentation files (`secret_redaction.py` is code that
  *handles* secrets, not a credential file).

**Not excluded:** public SSH keys (`*.pub`), public certificates, template /
example fixtures (`*.example`, `*.sample`, `*.template`, `*.tmpl`, `*.dist`).

**Overrides** (`exclude_secret_patterns` config key): a group slug (e.g.
`key_material_directories`) disables that whole group; the legacy `*secret*`
token disables the broad-basename and secret-store groups; any other glob is a
per-pattern allow (a matching file is never treated as secret).

When a secret file is detected, a warning is included in the indexing response.
Secret files are never stored in the index or cached content directory.

---

## File Size Limits

* **Default maximum:** 500 KB per file (configurable via `max_file_size`).
* Files exceeding the limit are skipped during discovery.
* A configurable **file count limit** (default: 500 files) prevents runaway indexing of extremely large repositories. Can be overridden using the `JCODEMUNCH_MAX_INDEX_FILES` environment variable.

---

## Binary File Detection

Binary files are excluded using a two-stage check:

1. **Extension-based detection** — common binary extensions (`.exe`, `.dll`, `.so`, `.png`, `.jpg`, `.zip`, `.wasm`, `.pyc`, `.class`, `.pdf`, `.db`, `.sqlite`, etc.).
2. **Content-based detection** — files containing null bytes within the first 8 KB are treated as binary and skipped, even if the extension suggests source code.

---

## Encoding Safety

* All file reads use `errors="replace"` to substitute invalid UTF-8 bytes with the Unicode replacement character (U+FFFD) instead of raising decode errors.
* Symbol content retrieval also uses `errors="replace"` to ensure safe decoding.
* Cached raw files are stored using UTF-8 encoding.

---

## Storage Safety

* Index storage defaults to `~/.code-index/`.
* The storage path can be overridden using the `CODE_INDEX_PATH` environment variable.
* Repository identifiers are derived from `{owner}-{name}`, preventing path injection in storage locations.
* Index files are stored as JSON and validated during load to ensure schema integrity.

---

## Release artifact signing

GitHub release artifacts (wheel + sdist) are signed with
[sigstore-python](https://github.com/sigstore/sigstore-python) via a
GitHub Actions workflow (`.github/workflows/sign-release.yml`) triggered
on `release.published`. The workflow uses GitHub's OIDC identity as the
signer, so verification ties an artifact back to the specific workflow
in this repository that signed it — no long-lived signing keys, no
external trust roots beyond the Sigstore public-good infrastructure.

**Verifying a release** (Sigstore v3 bundle format, `.sigstore.json`):

```bash
TAG=v1.108.22  # or whichever release you want to verify
WHEEL=jcodemunch_mcp-${TAG#v}-py3-none-any.whl
BASE="https://github.com/jgravelle/jcodemunch-mcp/releases/download/${TAG}"

curl -L -o "${WHEEL}" "${BASE}/${WHEEL}"
curl -L -o "${WHEEL}.sigstore.json" "${BASE}/${WHEEL}.sigstore.json"

python -m pip install sigstore
python -m sigstore verify github \
    --bundle "${WHEEL}.sigstore.json" \
    --repository jgravelle/jcodemunch-mcp \
    --workflow-name "Sign release artifacts" \
    "${WHEEL}"
```

The trust shape is the same one PyPI's PEP 740 attestation pipeline uses:
the workflow runs in GitHub Actions, presents an OIDC identity claim to
Sigstore's transparency log, and the signature is recoverable from the
log via the bundle. Forward-only — releases prior to the signing
workflow's introduction don't carry signatures and aren't going to be
retroactively resigned.

---

## Files this server treats as security-sensitive

The following user-writable files participate in the server's trust chain. A
process that can write any of them can influence the behavior of every
subsequent MCP session: prompt context the agent sees, tool descriptions, hook
commands, and which MCP server gets launched. Endpoint-management teams and
hardened install templates should treat them with the same care as any other
piece of developer configuration that steers an AI agent.

* `~/.code-index/config.jsonc` — global server configuration. Settings here
  influence tool tier visibility, language gating, secret-pattern lists, and
  per-tool description overrides.
* `~/.code-index/` and everything under it — the symbol index, the
  optional telemetry SQLite, the bundled-encoder model directory, and the
  serialized session journal. Bodies cached here are a second copy of every
  indexed source file.
* `./.jcodemunch.jsonc` (per-project) — same key shape as the global
  config, scoped to the directory it lives in. Overrides only those keys
  it sets.
* `~/.claude/CLAUDE.md`, `./CLAUDE.md`, `AGENTS.md`,
  `.cursor/rules/jcodemunch.mdc`, `.windsurfrules` — agent-policy files
  that `jcodemunch-mcp init` may write or modify, with consent. Each is
  rendered into the agent's prompt at session start by the corresponding
  client.
* `~/.claude/settings.json` (PreToolUse / PostToolUse / PreCompact /
  TaskCompleted / SubagentStart / WorktreeCreate / WorktreeRemove hooks)
  — `init` registers hook commands here so Claude Code auto-reindexes
  after edits and surfaces session diagnostics. The hook commands run
  every relevant tool call in the host agent.
* `.github/hooks/hooks.json` — analogous hook surface for GitHub Copilot
  CLI / cloud agent flows.
* Generated MCP client config files (paths depend on which clients are
  installed): `~/Library/Application Support/Claude/claude_desktop_config.json`
  (macOS Claude Desktop), `%APPDATA%\Claude\claude_desktop_config.json`
  (Windows Claude Desktop), `~/.cursor/mcp.json`, `~/.continue/config.json`,
  and the project-scope `.mcp.json` written by `claude mcp add`. Each
  contains the command line Claude / Cursor / Continue spawn to launch
  the MCP server.

File-integrity monitoring at the endpoint level (SentinelOne, Tanium, etc.)
applied to these paths is a reasonable defense-in-depth control in any
managed-endpoint deployment.

---

## Persistent processes installed by `watch-install`

`jcodemunch-mcp watch-install` registers a login-time service that watches
indexed directories for filesystem changes and reindexes incrementally. This
is opt-in and reversible (`watch-uninstall`) but appears in endpoint hunts
that enumerate startup items, so document it as expected when the service is
present:

* **Linux (systemd user units):** `~/.config/systemd/user/jcodemunch-watch.service`.
  Enabled with `systemctl --user enable --now jcodemunch-watch.service`.
* **macOS (launchd LaunchAgent):**
  `~/Library/LaunchAgents/us.gravelle.jcodemunch-watch.plist`. Loaded with
  `launchctl bootstrap gui/$UID <plist>`.
* **Windows (Task Scheduler entry):** task named `jcodemunch-watch` under
  the current user, configured to run at logon.

The service runs `jcodemunch-mcp watch-all`, which performs no network I/O
and only writes back to the per-repo SQLite stores under `~/.code-index/`.

---

## Cache integrity verification modes

`get_symbol_source(verify=True)` hashes the retrieved source and compares
against the content hash stored in the index. Both values are derived from
the local cache directory, so the default verification is self-referential:
a coherent tamper of `~/.code-index/<repo>/` is durably trusted after
the tamper. Treat the cache directory accordingly — see the security-sensitive
files section above for why it's worth file-integrity monitoring.

Externally-attested verification is available via the
`verify_against="git_sha"` parameter on `get_symbol_source`: when set, the
cached source is compared against the working-tree git HEAD slice of the
same file, not against the cache's own stored hash. The response includes
a `git_sha_verification` field with one of:

- `git_sha_match` — the cached source matches the HEAD slice.
- `git_sha_mismatch` — the file exists in HEAD but the slice differs.
- `git_unavailable` — the file isn't in HEAD, git is unreachable, or the
  source isn't a git working tree.

Default remains `verify_against="cache"` for back-compat. For
managed-endpoint or supply-chain-conscious deployments where cache
integrity matters, the `git_sha` mode is the externally-attested signal;
the `cache` mode alone is best read as "the cache is internally
consistent," not "the cache matches the upstream source."

---

## Telemetry Data Locality

The performance and ranking telemetry introduced in v1.74.0–v1.80.0 is
**local-only** and **opt-in**:

* `~/.code-index/telemetry.db` (`tool_calls`, `ranking_events`) is written
  only when `perf_telemetry_enabled: true` (or `JCODEMUNCH_PERF_TELEMETRY=1`).
  Default is **disabled** — the in-memory latency ring is always tracked
  but no row touches disk.
* `~/.code-index/tuning.jsonc` (per-repo retrieval-weight overrides) is
  written only by an explicit `tune_weights` invocation.
* `~/.code-index/embed_canary.json` (16-string drift canary) is written
  only by an explicit `check_embedding_drift(capture=true)` invocation.
* No telemetry is sent over the network. The community token-savings
  counter (`share_savings`) is unrelated and only sends an integer
  delta plus an anonymous UUID — never query strings, paths, or repo
  names. Disable with `JCODEMUNCH_SHARE_SAVINGS=0`.
* Stored ranking events include the **literal query string** (truncated
  result-id list, no source code). Treat the storage path with the same
  care as any local source you index.

---

## Summary of Controls

| Control                   | Location                       | Default                     |
| ------------------------- | ------------------------------ | --------------------------- |
| Path traversal validation | `security.validate_path()`     | Always enabled              |
| Symlink escape protection | `security.is_symlink_escape()` | Symlinks skipped by default |
| Secret file exclusion     | `security.is_secret_file()`    | Always enabled              |
| Binary file detection     | `security.is_binary_file()`    | Always enabled              |
| File size limit           | File discovery pipeline        | 500 KB                      |
| File count limit          | File discovery pipeline        | 500 files                   |
| `.gitignore` respect      | Indexing pipeline              | Enabled                     |
| UTF-8 safe decode         | All file reads                 | `errors="replace"`          |
| Perf telemetry sink       | `perf_telemetry_enabled`       | **Disabled** (opt-in)       |
| Ranking ledger storage    | `perf_telemetry_enabled`       | **Disabled** (opt-in)       |
| Tuning overrides          | Explicit `tune_weights` call   | None until invoked          |
| Embedding canary          | Explicit `check_embedding_drift` call | None until invoked   |
