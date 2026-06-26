"""``observatory`` — public OSS code-health observatory pipeline.

todo.md item #7. Periodically clones a curated list of OSS repos,
indexes each one, runs ``get_repo_health``, and writes static HTML +
JSON artifacts that an external host (Mac Mini, fly.io, GitHub Pages,
S3, anywhere) can serve as ``/report/<repo>``.

Hosting deliberately decoupled. The pipeline produces static files; the
hosting decision is "point a CDN at this directory" and shipping is
unblocked while that decision matures.

Pipeline shape:
    config (list of repos) -> for each:
        clone-or-update -> index_folder -> get_repo_health
                                       -> append to history.json
                                       -> render repo landing HTML
                                       -> render RSS feed
    -> render top-level index.html (leaderboard sorted by grade)

State lives entirely in the output directory:
    <output>/
        index.html                    # leaderboard
        index.json                    # machine-readable leaderboard
        feed.xml                      # cross-repo RSS
        <owner>--<repo>/
            index.html                # per-repo landing page
            history.json              # newest-first list of runs
            feed.xml                  # per-repo RSS
            radar.svg                 # current radar as static SVG

A run is `{timestamp, sha, grade, composite, axes, summary}`. Every run
appends to history (capped at 52 entries — a year of weekly runs);
re-runs at the same SHA are no-ops, so weekly cron over a quiet repo
is cheap.
"""

from __future__ import annotations

import datetime as _dt
import json
import logging
import shutil
import subprocess
import time
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

logger = logging.getLogger(__name__)


_DEFAULT_HISTORY_CAP = 52  # ~1 year of weekly runs


@dataclass
class RepoConfig:
    """One entry in the observatory config file."""
    url: str
    label: str = ""           # display name; defaults to <owner>/<repo>
    branch: Optional[str] = None
    languages: Optional[list[str]] = None  # for filtering languages
    # Optional one-line note shown on the repo's landing page (e.g.
    # "well-known web framework"). Free-form; safe to leave blank.
    blurb: str = ""


@dataclass
class ObservatoryConfig:
    repos: list[RepoConfig]
    output_dir: Path
    workdir: Path = field(default_factory=lambda: Path(".observatory_work"))
    history_cap: int = _DEFAULT_HISTORY_CAP


def load_config(path: Path) -> ObservatoryConfig:
    """Load and validate an observatory config from a JSON file."""
    raw = json.loads(path.read_text(encoding="utf-8"))
    repos = [RepoConfig(**r) for r in raw["repos"]]
    output_dir = Path(raw["output_dir"]).expanduser().resolve()
    workdir = Path(raw.get("workdir", ".observatory_work")).expanduser().resolve()
    history_cap = int(raw.get("history_cap", _DEFAULT_HISTORY_CAP))
    return ObservatoryConfig(
        repos=repos,
        output_dir=output_dir,
        workdir=workdir,
        history_cap=history_cap,
    )


def repo_slug(url: str) -> str:
    """Stable slug for a repo URL: github.com/owner/name -> owner--name."""
    s = url.strip().rstrip("/")
    if s.endswith(".git"):
        s = s[:-4]
    parts = [p for p in s.replace(":", "/").split("/") if p]
    if len(parts) < 2:
        return s.replace("/", "--")
    return f"{parts[-2]}--{parts[-1]}"


def _run_git(args: list[str], cwd: Path, *, timeout: int = 60) -> tuple[int, str, str]:
    try:
        r = subprocess.run(
            ["git", *args],
            cwd=str(cwd),
            capture_output=True,
            text=True,
            timeout=timeout,
            stdin=subprocess.DEVNULL,
        )
        return r.returncode, r.stdout.strip(), r.stderr.strip()
    except (subprocess.TimeoutExpired, FileNotFoundError, OSError) as e:
        return -1, "", str(e)


def clone_or_update(url: str, dest: Path, branch: Optional[str] = None) -> Optional[str]:
    """Ensure ``dest`` contains a checkout of ``url``; return current HEAD SHA.

    Shallow clone (depth=1) is sufficient for indexing — we don't need
    history, just the current tree. Subsequent runs do a fast-forward
    fetch + reset to the latest tip.
    """
    dest.parent.mkdir(parents=True, exist_ok=True)
    if not (dest / ".git").is_dir():
        if dest.exists():
            shutil.rmtree(dest)
        clone_args = ["clone", "--depth=1"]
        if branch:
            clone_args += ["--branch", branch]
        clone_args += [url, str(dest)]
        rc, _, err = _run_git(clone_args, cwd=dest.parent, timeout=300)
        if rc != 0:
            logger.warning("clone %s failed: %s", url, err)
            return None
    else:
        # Fast-forward update.
        ref = branch or "HEAD"
        rc, _, err = _run_git(["fetch", "--depth=1", "origin", ref], cwd=dest, timeout=120)
        if rc != 0:
            logger.warning("fetch %s failed: %s", url, err)
            return None
        rc, _, err = _run_git(["reset", "--hard", "FETCH_HEAD"], cwd=dest, timeout=30)
        if rc != 0:
            logger.warning("reset %s failed: %s", url, err)
            return None

    rc, sha, _ = _run_git(["rev-parse", "HEAD"], cwd=dest, timeout=10)
    return sha if rc == 0 else None


def index_and_health(repo_path: Path, *, storage_path: Optional[str] = None) -> Optional[dict]:
    """Index ``repo_path`` and return its health response (with radar)."""
    from .index_folder import index_folder
    from .get_repo_health import get_repo_health

    # Index quietly; failures bubble up as None.
    try:
        idx_result = index_folder(
            str(repo_path),
            use_ai_summaries=False,
            storage_path=storage_path,
        )
        if idx_result.get("error") or not idx_result.get("success"):
            logger.warning("index failed for %s: %s", repo_path, idx_result.get("error") or idx_result)
            return None
    except Exception:
        logger.exception("index_folder crashed on %s", repo_path)
        return None

    repo_id = idx_result.get("repo")
    if not repo_id or "/" not in repo_id:
        logger.warning("index_folder returned no repo identifier for %s (got %r)", repo_path, repo_id)
        return None

    try:
        health = get_repo_health(repo=repo_id, storage_path=storage_path)
        if health.get("error"):
            logger.warning("get_repo_health error on %s: %s", repo_path, health["error"])
            return None
        return health
    except Exception:
        logger.exception("get_repo_health crashed on %s", repo_path)
        return None


def history_path(output_dir: Path, slug: str) -> Path:
    return output_dir / slug / "history.json"


def load_history(output_dir: Path, slug: str) -> list[dict]:
    path = history_path(output_dir, slug)
    if not path.exists():
        return []
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
        return data if isinstance(data, list) else []
    except (OSError, json.JSONDecodeError):
        return []


def append_run(
    output_dir: Path,
    slug: str,
    sha: str,
    health: dict,
    *,
    cap: int = _DEFAULT_HISTORY_CAP,
) -> dict:
    """Append a run record to the per-repo history file. Idempotent on SHA.

    Returns the new run record (or the existing top entry if SHA matches).
    """
    history = load_history(output_dir, slug)
    if history and history[0].get("sha") == sha:
        return history[0]

    radar = health.get("radar") or {}
    radar_axes = radar.get("axes") or {}
    # Phase 7: presence of the runtime_coverage axis signals whether the
    # repo has any ingested runtime evidence at all. The axis is omitted
    # for repos that haven't run import_runtime_signal — never penalised,
    # just flagged so the leaderboard can distinguish empirical scores
    # from purely-static ones.
    has_runtime_evidence = "runtime_coverage" in radar_axes
    record = {
        "timestamp": _dt.datetime.now(_dt.timezone.utc).isoformat(),
        "sha": sha,
        "summary": health.get("summary", ""),
        "grade": radar.get("grade", "?"),
        "composite": float(radar.get("composite", 0.0) or 0.0),
        "total_files": int(health.get("total_files", 0) or 0),
        "total_symbols": int(health.get("total_symbols", 0) or 0),
        "runtime_evidence": has_runtime_evidence,
        "axes": {
            axis: round(float(d.get("score", 0.0) or 0.0), 1)
            for axis, d in radar_axes.items()
        },
    }
    history.insert(0, record)
    if len(history) > cap:
        history = history[:cap]

    path = history_path(output_dir, slug)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(history, indent=2) + "\n", encoding="utf-8")
    return record


def run_pipeline(config: ObservatoryConfig) -> dict:
    """End-to-end pipeline: clone+index+score every repo, write artifacts.

    Returns a summary dict with per-repo status; the actual outputs are
    static files in ``config.output_dir``.
    """
    from . import observatory_render as render

    config.workdir.mkdir(parents=True, exist_ok=True)
    config.output_dir.mkdir(parents=True, exist_ok=True)

    # Each observatory run uses an isolated index storage so it doesn't
    # pollute the user's main ~/.code-index/ with public OSS repos.
    storage_path = str(config.workdir / ".index")

    summaries: list[dict] = []
    for repo in config.repos:
        slug = repo_slug(repo.url)
        repo_path = config.workdir / "checkouts" / slug
        label = repo.label or "/".join(repo.url.rstrip("/").split("/")[-2:])

        t0 = time.perf_counter()
        sha = clone_or_update(repo.url, repo_path, branch=repo.branch)
        if not sha:
            summaries.append({"slug": slug, "label": label, "status": "clone_failed", "url": repo.url})
            continue

        health = index_and_health(repo_path, storage_path=storage_path)
        if not health:
            summaries.append({"slug": slug, "label": label, "status": "health_failed", "url": repo.url, "sha": sha})
            continue

        record = append_run(config.output_dir, slug, sha, health, cap=config.history_cap)
        history = load_history(config.output_dir, slug)
        render.render_repo_page(config.output_dir, slug, label, repo, history, health)

        elapsed_s = round(time.perf_counter() - t0, 1)
        summaries.append({
            "slug": slug,
            "label": label,
            "status": "ok",
            "url": repo.url,
            "sha": sha,
            "grade": record["grade"],
            "composite": record["composite"],
            "runtime_evidence": record.get("runtime_evidence", False),
            "elapsed_s": elapsed_s,
        })

    render.render_index_page(config.output_dir, summaries)
    render.render_index_feed(config.output_dir, summaries)

    return {
        "total": len(summaries),
        "ok": sum(1 for s in summaries if s["status"] == "ok"),
        "failed": sum(1 for s in summaries if s["status"] != "ok"),
        "summaries": summaries,
    }
