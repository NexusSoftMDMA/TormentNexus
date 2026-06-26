"""First-launch version-drift probe.

When the server starts, compare the current package version against
``~/.code-index/last_seen_version``. On mismatch, write a one-line stderr
hint pointing at the release notes and update the file. On first launch
(file absent), write the current version with no message.

Failure modes are silent — this is a UX nicety, not a critical path. Any
OSError, permission issue, or unexpected file content drops through to
no-op so the server starts normally.

Disable with ``JCODEMUNCH_NO_VERSION_HINT=1``.
"""

from __future__ import annotations

import logging
import os
import sys
from pathlib import Path

logger = logging.getLogger(__name__)

_LAST_SEEN_FILENAME = "last_seen_version"
_RELEASE_URL_FMT = "https://github.com/jgravelle/jcodemunch-mcp/releases/tag/v{version}"


def _storage_root() -> Path:
    """Same root the rest of the package uses for cached state."""
    env = os.environ.get("CODE_INDEX_PATH")
    if env:
        return Path(env)
    return Path.home() / ".code-index"


def _current_version() -> str | None:
    """Return the running package version, or None if undeterminable."""
    try:
        from importlib.metadata import version as _v
        return _v("jcodemunch-mcp")
    except Exception:
        logger.debug("could not determine running version", exc_info=True)
        return None


def check_and_announce() -> None:
    """Compare current vs. last-seen version; emit one-line stderr hint on drift.

    Called once at server startup. Silent on first launch (no prior
    version file) and on any OS-level failure.
    """
    if os.environ.get("JCODEMUNCH_NO_VERSION_HINT", "0") in ("1", "true", "yes"):
        return

    current = _current_version()
    if not current:
        return

    root = _storage_root()
    try:
        root.mkdir(parents=True, exist_ok=True)
    except OSError:
        logger.debug("could not create storage root for version probe", exc_info=True)
        return

    seen_path = root / _LAST_SEEN_FILENAME

    prior: str | None = None
    if seen_path.exists():
        try:
            prior = seen_path.read_text(encoding="utf-8").strip()
        except OSError:
            logger.debug("could not read last_seen_version", exc_info=True)
            return

    if prior == current:
        return  # No drift, no message.

    try:
        seen_path.write_text(current, encoding="utf-8")
    except OSError:
        logger.debug("could not write last_seen_version", exc_info=True)
        # Continue anyway — we can still print the hint.

    if prior is None:
        # First launch — write the file silently, no announcement.
        return

    release_url = _RELEASE_URL_FMT.format(version=current)
    print(
        f"[jcodemunch-mcp] upgraded {prior} → {current} — release notes: {release_url}",
        file=sys.stderr,
        flush=True,
    )
