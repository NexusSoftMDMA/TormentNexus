"""Update rollback state for taOS.

Before every update, the updater records the exact branch + commit it is leaving
so a later ``taos rollback`` can restore BOTH (the previous version and the
previous branch, even if both changed). The record is written as a tiny
shell-sourceable file so ``scripts/rollback.sh`` can read it with no Python and
no dashboard, which is the whole point: rollback must work when an update has
broken the app.

File: ``<project_dir>/.taos-rollback`` (single record, overwritten each update).
"""

from __future__ import annotations

from pathlib import Path

ROLLBACK_FILE = ".taos-rollback"


def _shq(value: str) -> str:
    """Single-quote a value so the file stays safe to `source` in bash."""
    return "'" + str(value).replace("'", "'\\''") + "'"


def record_pre_update(project_dir, *, branch: str, sha: str, ts: int) -> Path:
    """Write the pre-update branch + commit so a rollback can restore both.

    Overwrites any prior record: rollback targets the state immediately before
    the most recent update, which is the one a user would want to undo.
    """
    path = Path(project_dir) / ROLLBACK_FILE
    path.write_text(
        "# taOS rollback target -- the branch + commit the last update left.\n"
        "# Shell-sourceable on purpose so scripts/rollback.sh needs no Python.\n"
        f"prev_branch={_shq(branch)}\n"
        f"prev_sha={_shq(sha)}\n"
        f"prev_ts={_shq(ts)}\n"
    )
    return path


def read_rollback_target(project_dir) -> dict | None:
    """Read the recorded rollback target, or None if there is no record.

    Returns ``{"branch": str, "sha": str, "ts": str}``. Parses the simple
    ``key='value'`` lines without sourcing (so it is safe to call on any input).
    """
    path = Path(project_dir) / ROLLBACK_FILE
    if not path.is_file():
        return None
    out: dict[str, str] = {}
    for line in path.read_text().splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, _, raw = line.partition("=")
        val = raw.strip()
        if len(val) >= 2 and val[0] == val[-1] == "'":
            val = val[1:-1].replace("'\\''", "'")
        out[key.strip()] = val
    if "prev_branch" not in out or "prev_sha" not in out:
        return None
    return {"branch": out["prev_branch"], "sha": out["prev_sha"], "ts": out.get("prev_ts", "")}
