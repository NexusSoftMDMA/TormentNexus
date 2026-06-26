"""Hook event handler — receives Claude Code worktree events, creates/removes
git worktrees, and records state to a JSONL manifest."""

import json
import os
import shutil
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path


def default_manifest_path() -> Path:
    """Return manifest path, respecting CODE_INDEX_PATH."""
    storage = os.environ.get("CODE_INDEX_PATH", str(Path.home() / ".code-index"))
    return Path(storage) / "jcodemunch-worktrees.jsonl"


# Legacy location — checked for migration.
_LEGACY_MANIFEST_PATH = Path.home() / ".claude" / "jcodemunch-worktrees.jsonl"


def _migrate_manifest(new_path: Path) -> None:
    """Move legacy manifest from ~/.claude/ to the new location if it exists."""
    if _LEGACY_MANIFEST_PATH.is_file() and not new_path.is_file():
        new_path.parent.mkdir(parents=True, exist_ok=True)
        shutil.move(str(_LEGACY_MANIFEST_PATH), str(new_path))


def _get_worktree_base() -> str:
    """Load worktree_base_path from config.jsonc (empty string = use default)."""
    try:
        from jcodemunch_mcp.config import load_config, get as config_get
        load_config()
        return config_get("worktree_base_path", "")
    except Exception:
        return ""


def _resolve_worktree_path(cwd: str, name: str) -> str:
    """Determine the worktree path from cwd + name, respecting config."""
    base = _get_worktree_base()
    if base:
        return str(Path(base).expanduser().resolve() / name)
    return str(Path(cwd) / ".claude" / "worktrees" / name)


def _resolve_main_repo(cwd: str) -> str:
    """Return the main repo root, even when *cwd* is inside a worktree.

    Uses ``git rev-parse --git-common-dir`` to find the shared ``.git``
    directory, then takes its parent.  Falls back to *cwd* on any error.
    """
    try:
        result = subprocess.run(
            ["git", "-C", cwd, "rev-parse", "--path-format=absolute",
             "--git-common-dir"],
            capture_output=True, text=True,
            stdin=subprocess.DEVNULL,  # don't inherit the hook's stdin pipe (Windows git wrapper deadlock)
        )
        if result.returncode == 0:
            git_common = result.stdout.strip()
            return str(Path(git_common).parent)
    except Exception:
        pass
    return cwd


def _append_manifest(event_type: str, resolved: str, manifest_path: Path) -> None:
    """Append a single event line to the JSONL manifest."""
    entry = {
        "event": event_type,
        "path": resolved,
        "ts": datetime.now(timezone.utc).isoformat(),
    }
    manifest_path.parent.mkdir(parents=True, exist_ok=True)
    with open(manifest_path, "a") as f:
        f.write(json.dumps(entry) + "\n")
        f.flush()


def handle_hook_event(event_type: str, manifest_path: Path | None = None) -> None:
    """Handle a WorktreeCreate or WorktreeRemove event from Claude Code.

    For 'create': determines path, runs ``git worktree add``, records to
    manifest, prints resolved path to stdout.

    For 'remove': determines path, runs ``git worktree remove``, records
    to manifest.

    When ``worktreePath`` is provided directly for a 'create' event, the
    caller already owns the worktree — only manifest recording is performed.
    For 'remove', the worktree is still torn down even when the path is
    provided directly.

    Called by Claude Code hooks via:
        jcodemunch-mcp hook-event create
        jcodemunch-mcp hook-event remove
    """
    if manifest_path is None:
        manifest_path = default_manifest_path()
        _migrate_manifest(manifest_path)

    try:
        payload = json.load(sys.stdin)
    except json.JSONDecodeError as exc:
        print(f"ERROR: invalid JSON on stdin: {exc}", file=sys.stderr)
        sys.exit(1)
    except Exception as exc:
        print(f"ERROR: failed to read stdin: {exc}", file=sys.stderr)
        sys.exit(1)

    cwd = payload.get("cwd", "")
    name = payload.get("name", "")

    # Accept worktreePath / worktree_path directly (Claude Code sends this
    # for WorktreeRemove).  For 'create' the caller already owns the
    # worktree — just record the manifest.  For 'remove' we still need to
    # tear down the worktree and branch.
    legacy_path = payload.get("worktreePath") or payload.get("worktree_path")

    if legacy_path and event_type == "create":
        resolved = str(Path(legacy_path).resolve())
        _append_manifest(event_type, resolved, manifest_path)
        print(resolved, flush=True)
        return

    if legacy_path:
        # 'remove' with an explicit path — use it directly.
        resolved = str(Path(legacy_path).resolve())
        # Derive name from the last path component for branch cleanup.
        if not name:
            name = Path(resolved).name
    elif cwd and name:
        worktree_path = _resolve_worktree_path(cwd, name)
        resolved = str(Path(worktree_path).resolve())
    else:
        print("ERROR: need cwd+name or worktreePath in stdin payload", file=sys.stderr)
        sys.exit(1)

    if event_type == "create":
        Path(resolved).parent.mkdir(parents=True, exist_ok=True)
        branch_name = f"worktree-{name}"
        result = subprocess.run(
            ["git", "-C", cwd, "worktree", "add", resolved, "-b", branch_name],
            capture_output=True,
            text=True,
            stdin=subprocess.DEVNULL,  # don't inherit the hook's stdin pipe (Windows git wrapper deadlock)
        )
        if result.returncode != 0:
            print(f"ERROR: git worktree add failed: {result.stderr.strip()}", file=sys.stderr)
            sys.exit(1)

    elif event_type == "remove":
        # cwd may be the worktree itself — resolve to the main repo so
        # git worktree remove doesn't run from inside the tree it's deleting.
        repo_root = _resolve_main_repo(cwd) if cwd else _resolve_main_repo(resolved)
        result = subprocess.run(
            ["git", "-C", repo_root, "worktree", "remove", resolved, "--force"],
            capture_output=True,
            text=True,
            stdin=subprocess.DEVNULL,  # don't inherit the hook's stdin pipe (Windows git wrapper deadlock)
        )
        # Non-fatal: worktree may already be gone.
        if result.returncode != 0:
            print(f"WARNING: git worktree remove failed: {result.stderr.strip()}", file=sys.stderr)

        # Clean up the branch created during worktree add.
        branch_name = f"worktree-{name}"
        subprocess.run(
            ["git", "-C", repo_root, "branch", "-D", branch_name],
            capture_output=True,
            text=True,
            stdin=subprocess.DEVNULL,  # don't inherit the hook's stdin pipe (Windows git wrapper deadlock)
        )
        # Branch deletion is best-effort — no error if it doesn't exist.

    _append_manifest(event_type, resolved, manifest_path)

    # Claude Code reads stdout to get the worktree path.
    # flush=True ensures output is not buffered (e.g. when run via uvx).
    print(resolved, flush=True)


def read_manifest(manifest_path: Path | None = None) -> set[str]:
    """Read the JSONL manifest and return the set of currently active worktree paths.

    Replays all events in order: a 'create' adds a path, a 'remove' removes it.
    """
    if manifest_path is None:
        manifest_path = default_manifest_path()
        _migrate_manifest(manifest_path)

    active: dict[str, bool] = {}
    if not manifest_path.is_file():
        return set()

    with open(manifest_path) as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                entry = json.loads(line)
            except json.JSONDecodeError:
                continue
            path = entry.get("path")
            event = entry.get("event")
            if not path or event not in ("create", "remove"):
                continue
            active[path] = event == "create"

    return {p for p, is_active in active.items() if is_active}
