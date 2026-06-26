"""Per-symbol freshness classification (v1.77.0).

Three buckets:
  * ``fresh``               — index SHA matches HEAD AND the file mtime is
                              not newer than the index timestamp.
  * ``edited_uncommitted``  — index SHA matches HEAD but the on-disk file
                              has been edited since indexing (mtime newer
                              than indexed_at).
  * ``stale_index``         — the whole index lags behind: index SHA does
                              not match the current git HEAD.

The probe caches per-call git HEAD lookup and per-file mtime stats so
classifying many symbols in one tool call is cheap.
"""

from __future__ import annotations

import logging
import subprocess
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional

logger = logging.getLogger(__name__)


_FRESH = "fresh"
_EDITED = "edited_uncommitted"
_STALE = "stale_index"


def _parse_iso(ts: str) -> Optional[float]:
    """Parse the ISO timestamp recorded in the index. Returns Unix epoch
    seconds (float) or None on parse failure."""
    if not ts:
        return None
    try:
        # tolerate trailing Z
        if ts.endswith("Z"):
            ts = ts[:-1] + "+00:00"
        return datetime.fromisoformat(ts).astimezone(timezone.utc).timestamp()
    except Exception:
        return None


# Process-wide cache of resolved git HEAD per source root (B1). The subprocess
# is re-run only when a cheap stat-based signature of the refs that move with
# HEAD changes; when no signature can be computed (exotic layouts) a short TTL
# bounds reuse so a burst of tool calls shares one `git rev-parse` rather than
# spawning one each. key -> (signature, sha, monotonic_ts).
_HEAD_CACHE_TTL_S = 2.0
_head_cache: dict[str, tuple[Optional[tuple], Optional[str], float]] = {}


def _clear_head_cache() -> None:
    """Test hook: drop all cached HEAD lookups."""
    _head_cache.clear()


def _resolve_git_dir(source_root: Path) -> Optional[Path]:
    """Return the .git directory for *source_root*, resolving worktree/submodule
    `.git` files (``gitdir: <path>``). None when it isn't a git repo."""
    dotgit = source_root / ".git"
    try:
        if dotgit.is_dir():
            return dotgit
        if dotgit.is_file():
            text = dotgit.read_text(encoding="utf-8", errors="ignore").strip()
            if text.startswith("gitdir:"):
                p = Path(text[len("gitdir:"):].strip())
                return p if p.is_absolute() else (source_root / p).resolve()
    except OSError:
        return None
    return None


def _head_signature(git_dir: Path) -> Optional[tuple]:
    """Stat-based signature that changes whenever HEAD's commit moves.

    Covers ordinary commits (loose ref + reflog), ref packing (packed-refs),
    and branch switch / detach (HEAD content). Returns None when nothing could
    be stat'd, signalling the caller to fall back to TTL-bounded caching.
    """
    paths = [
        git_dir / "HEAD",
        git_dir / "packed-refs",
        git_dir / "logs" / "HEAD",
    ]
    try:
        head_txt = (git_dir / "HEAD").read_text(encoding="utf-8", errors="ignore").strip()
    except OSError:
        head_txt = ""
    if head_txt.startswith("ref:"):
        ref = head_txt[4:].strip()
        paths.append(git_dir / ref)
        # Worktrees keep shared refs (loose + packed) in the common dir.
        try:
            commondir = (git_dir / "commondir").read_text(encoding="utf-8", errors="ignore").strip()
        except OSError:
            commondir = ""
        if commondir:
            base = (git_dir / commondir).resolve()
            paths.append(base / ref)
            paths.append(base / "packed-refs")
    sig: list[tuple[str, Optional[int]]] = []
    found = False
    for p in paths:
        try:
            sig.append((str(p), p.stat().st_mtime_ns))
            found = True
        except OSError:
            sig.append((str(p), None))
    return tuple(sig) if found else None


def _git_head_uncached(source_root: Path) -> Optional[str]:
    try:
        result = subprocess.run(
            ["git", "rev-parse", "HEAD"],
            cwd=str(source_root),
            capture_output=True,
            text=True,
            timeout=2,
            stdin=subprocess.DEVNULL,
        )
        if result.returncode == 0:
            return result.stdout.strip()
    except Exception:
        logger.debug("git rev-parse HEAD failed at %s", source_root, exc_info=True)
    return None


def _git_head(source_root: Path) -> Optional[str]:
    """Cached ``git rev-parse HEAD``. Re-runs the subprocess only when the HEAD
    signature changes, or — when no signature is available — at most once per
    TTL window per repo. Always safe: a cache miss just recomputes."""
    key = str(source_root)
    git_dir = _resolve_git_dir(source_root)
    sig = _head_signature(git_dir) if git_dir else None
    now = time.monotonic()

    cached = _head_cache.get(key)
    if cached is not None:
        c_sig, c_sha, c_ts = cached
        if sig is not None and c_sig is not None and sig == c_sig:
            return c_sha
        if sig is None and c_sig is None and (now - c_ts) < _HEAD_CACHE_TTL_S:
            return c_sha

    sha = _git_head_uncached(source_root)
    _head_cache[key] = (sig, sha, now)
    return sha


class FreshnessProbe:
    """Per-call freshness classifier.

    Construct once per tool invocation, then call ``classify(file_path)``
    for each returned symbol's file. The probe holds:
      * ``index_sha``  — the SHA stored at index time.
      * ``current_sha`` — fresh git HEAD (lazy, cached).
      * ``indexed_ts`` — Unix epoch of the index timestamp.
      * ``mtime_cache`` — per-file mtime memo (str → float | None).
    """

    def __init__(
        self,
        source_root: Optional[str],
        indexed_at: str,
        index_sha: Optional[str],
        *,
        current_sha: Optional[str] = None,
        file_mtimes: Optional[dict] = None,
    ):
        self._source_root = Path(source_root) if source_root else None
        self._index_sha = index_sha or None
        self._indexed_ts = _parse_iso(indexed_at)
        self._current_sha = current_sha  # may be None (lazy)
        self._current_sha_resolved = current_sha is not None
        self._mtime_cache: dict[str, Optional[float]] = {}
        # Per-file mtime recorded at index time (CodeIndex.file_mtimes is in
        # nanoseconds; convert to seconds). When available, comparison is
        # per-file rather than against a single index-wide indexed_at.
        self._indexed_mtimes_s: dict[str, float] = {}
        if file_mtimes:
            for path, ns in file_mtimes.items():
                try:
                    self._indexed_mtimes_s[path] = float(ns) / 1e9
                except (TypeError, ValueError):
                    pass

    @property
    def repo_is_stale(self) -> bool:
        """True iff index SHA differs from the live HEAD (and we know HEAD)."""
        cur = self._lazy_current_sha()
        if not cur or not self._index_sha:
            return False
        return cur != self._index_sha

    def _lazy_current_sha(self) -> Optional[str]:
        if self._current_sha_resolved:
            return self._current_sha
        if self._source_root and self._source_root.exists():
            self._current_sha = _git_head(self._source_root)
        self._current_sha_resolved = True
        return self._current_sha

    def _file_mtime(self, file_rel: str) -> Optional[float]:
        if file_rel in self._mtime_cache:
            return self._mtime_cache[file_rel]
        if not self._source_root:
            self._mtime_cache[file_rel] = None
            return None
        try:
            p = self._source_root / file_rel
            mtime = p.stat().st_mtime if p.exists() else None
        except OSError:
            mtime = None
        self._mtime_cache[file_rel] = mtime
        return mtime

    def classify(self, file_rel: str) -> str:
        """Return one of fresh / edited_uncommitted / stale_index."""
        if self.repo_is_stale:
            return _STALE
        if not file_rel:
            return _FRESH
        mtime_now = self._file_mtime(file_rel)
        if mtime_now is None:
            return _FRESH
        # Prefer per-file indexed mtime when available (more accurate than
        # the single index-wide indexed_at timestamp).
        per_file_indexed = self._indexed_mtimes_s.get(file_rel)
        if per_file_indexed is not None:
            if mtime_now > per_file_indexed + 1.0:
                return _EDITED
            return _FRESH
        if self._indexed_ts and mtime_now > self._indexed_ts + 1.0:
            return _EDITED
        return _FRESH

    def annotate(self, entries: list[dict], file_field: str = "file") -> list[dict]:
        """In-place ``_freshness`` annotation on a list of result entries.

        Returns the same list for chaining.
        """
        for entry in entries:
            if not isinstance(entry, dict):
                continue
            file_rel = entry.get(file_field) or ""
            entry["_freshness"] = self.classify(file_rel)
        return entries

    def summary(self, entries: list[dict]) -> dict:
        """Bucket-count summary of ``_freshness`` across entries."""
        counts = {_FRESH: 0, _EDITED: 0, _STALE: 0}
        for e in entries:
            if isinstance(e, dict):
                counts[e.get("_freshness", _FRESH)] = counts.get(
                    e.get("_freshness", _FRESH), 0
                ) + 1
        return {
            "fresh": counts.get(_FRESH, 0),
            "edited_uncommitted": counts.get(_EDITED, 0),
            "stale_index": counts.get(_STALE, 0),
            "repo_is_stale": self.repo_is_stale,
        }
