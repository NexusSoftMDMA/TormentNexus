# RELEASE_RESTORE.md — restoring normal PyPI publishing after the outage

Internal maintainer runbook. During the 2026-05 incident, PyPI publishing for
`jcodemunch-mcp` was unavailable (project under PyPI admin review; appeal filed
via security@pypi.org). The suite kept shipping through the **GitHub-release
wheel** channel plus `git+https` while the package name was unreachable on PyPI
(see the README banner and issue #308). This is the checklist to return to
normal PyPI publishing once access is restored. Public status / user-facing
checklist lives in #308; this file is the operational detail behind it.

## 0. How to tell it's restored

- pypi.org login shows no review/suspension notice.
- <https://pypi.org/project/jcodemunch-mcp/> loads, and the JSON API returns 200
  (during the outage it 404s): `curl -s -o /dev/null -w "%{http_code}" https://pypi.org/pypi/jcodemunch-mcp/json`
- `python -m pip index versions jcodemunch-mcp` lists versions (during the outage:
  "No matching distribution found").

## 1. Rotate credentials

API tokens were invalidated during the incident, so the first upload will 403
on the old token (`Invalid or non-existent authentication information`).

- Mint a fresh token: pypi.org -> Account settings -> API tokens (account-scoped,
  or project-scoped once the project itself is un-restricted).
- Update `~/.pypirc` (`[pypi]`, `username = __token__`, `password = pypi-...`).

## 2. Re-publish to PyPI

- Artifacts for v1.108.25 are already built in `dist/` (wheel + sdist). Build any
  newer versions cut during the outage: `python -m build`.
- Upload: `python -m twine upload dist/jcodemunch_mcp-1.108.25*` (plus any newer).
- Verify in a clean venv: `pip install jcodemunch-mcp==1.108.25` and
  `uvx jcodemunch-mcp`.

## 3. Revert docs + badges to the plain PyPI install

Files: `README.md`, `QUICKSTART.md`, `USER_GUIDE.md`, `CLAUDE.md`.

- **README + QUICKSTART** — re-point the one-click badges from the pinned GitHub
  wheel back to plain `uvx jcodemunch-mcp`:
  - VS Code / Insiders badge hrefs are `vscode:mcp/install?<urlencoded JSON>` with
    JSON `{"name":"jcodemunch","command":"uvx","args":["jcodemunch-mcp"]}`.
  - Cursor badge href is `...?name=jcodemunch&config=<base64 JSON>` with JSON
    `{"command":"uvx","args":["jcodemunch-mcp"]}`.
  - Remove the "plain `pip install jcodemunch-mcp` ... temporarily unavailable
    (#308)" Note breadcrumb under the badges.
  - Keep the `git+https` one-liner as the "from source / latest" alternative if
    desired (it is version-free and useful beyond the outage).
- **USER_GUIDE** — Antigravity config `args` back to `["jcodemunch-mcp"]`; delete
  the "...is the temporary PyPI-outage workaround" paragraph.
- **CLAUDE.md** — drop the "Shipped via GitHub-release wheel (PyPI ...)" qualifiers
  from current-state entries going forward.

Badge re-encode helper (confirm the exact pre-outage shape from git history first:
`git log --oneline -- README.md` then `git show <pre-outage-sha>:README.md`):

```python
import json, base64, urllib.parse
vscode = {"name": "jcodemunch", "command": "uvx", "args": ["jcodemunch-mcp"]}
cursor = {"command": "uvx", "args": ["jcodemunch-mcp"]}
print("vscode query:", urllib.parse.quote(json.dumps(vscode)))
print("cursor config:", base64.b64encode(json.dumps(cursor).encode()).decode())
```

## 4. Sibling packages

Uploads were blocked account-wide, so any `jdocmunch-mcp` / `jdatamunch-mcp`
releases prepared during the outage were never pushed to PyPI (their already
-published versions kept serving). Publish those now.

## 5. Close out

- Tick the restoration checklist in #308, post the resolution, close the issue.
- Update memory `project_pypi_suspension.md` (mark resolved or delete the file).
- GitHub-release wheels stay as a fallback channel but are no longer advertised
  as the primary install.

## Notes

- Keep the security@pypi.org appeal thread for the record.
- If the project is restored but the account is not (or vice versa), run only the
  applicable subset.
