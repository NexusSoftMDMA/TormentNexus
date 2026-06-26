"""Static-site rendering for the observatory pipeline.

Pure string templates — no Jinja, no JS framework, no runtime CSS
download. Output is plain HTML + inline SVG that can serve from any
static host (Mac Mini Caddy, GitHub Pages, S3, fly.io static).

Each rendering function takes already-prepared data and writes its
files. The pipeline (``observatory.py``) handles orchestration.
"""

from __future__ import annotations

import html
import json
import xml.sax.saxutils as _sax
from datetime import datetime, timezone
from pathlib import Path


_GRADE_COLOR = {
    "A": "#22c55e",
    "B": "#84cc16",
    "C": "#eab308",
    "D": "#f97316",
    "F": "#ef4444",
    "?": "#9ca3af",
}

_AXIS_LABELS = {
    "complexity":     "Complexity",
    "dead_code":      "Dead code",
    "cycles":         "Cycles",
    "coupling":       "Coupling",
    "test_gap":       "Test gap",
    "churn_surface":  "Churn surface",
}


_BASE_CSS = """
:root {
  color-scheme: light dark;
  --bg: #0b0d12;
  --fg: #e7eaf2;
  --muted: #94a3b8;
  --card: #131722;
  --border: #1f2733;
  --accent: #6366f1;
}
@media (prefers-color-scheme: light) {
  :root { --bg:#fafbfc; --fg:#0b0d12; --muted:#475569; --card:#ffffff; --border:#e5e7eb; --accent:#4f46e5; }
}
* { box-sizing: border-box; }
body {
  margin: 0; padding: 2rem 1rem; font: 15px/1.55 -apple-system, BlinkMacSystemFont, "Segoe UI", system-ui, sans-serif;
  background: var(--bg); color: var(--fg); max-width: 64rem; margin-left: auto; margin-right: auto;
}
h1 { font-size: 1.75rem; margin: 0 0 0.5rem; }
h2 { font-size: 1.15rem; margin: 1.5rem 0 0.5rem; }
a { color: var(--accent); text-decoration: none; }
a:hover { text-decoration: underline; }
.muted { color: var(--muted); }
.card { background: var(--card); border: 1px solid var(--border); border-radius: 0.5rem; padding: 1rem; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 0.4rem 0.6rem; border-bottom: 1px solid var(--border); text-align: left; }
th { font-weight: 600; color: var(--muted); }
.grade { display: inline-block; min-width: 2rem; padding: 0.1rem 0.4rem; border-radius: 0.25rem; font-weight: 600; text-align: center; }
.bar-track { height: 0.4rem; background: var(--border); border-radius: 0.2rem; overflow: hidden; }
.bar-fill { height: 100%; }
.axis-row td { padding: 0.3rem 0.6rem; border-bottom: 1px dashed var(--border); }
.footer { margin-top: 3rem; padding-top: 1rem; border-top: 1px solid var(--border); color: var(--muted); font-size: 0.875rem; }
.repo-list { display: grid; gap: 0.75rem; grid-template-columns: 1fr; }
@media (min-width: 720px) { .repo-list { grid-template-columns: 1fr 1fr; } }
.tile { display: block; padding: 1rem; background: var(--card); border: 1px solid var(--border); border-radius: 0.5rem; }
.tile:hover { border-color: var(--accent); }
.tile h3 { margin: 0 0 0.25rem; font-size: 1.05rem; }
.spark { vertical-align: middle; }
"""


def _grade_color(grade: str) -> str:
    return _GRADE_COLOR.get(grade, _GRADE_COLOR["?"])


def _composite_color(composite: float) -> str:
    if composite >= 90: return _GRADE_COLOR["A"]
    if composite >= 80: return _GRADE_COLOR["B"]
    if composite >= 70: return _GRADE_COLOR["C"]
    if composite >= 60: return _GRADE_COLOR["D"]
    return _GRADE_COLOR["F"]


def _esc(s) -> str:
    return html.escape(str(s))


def _format_iso_date(ts: str) -> str:
    """ISO timestamp -> 'YYYY-MM-DD HH:MMZ' for display."""
    if not ts:
        return ""
    try:
        dt = datetime.fromisoformat(ts.replace("Z", "+00:00"))
        return dt.strftime("%Y-%m-%d %H:%MZ")
    except ValueError:
        return ts


def render_axis_bar(label: str, score: float) -> str:
    """One axis row: name, bar, score."""
    score = max(0.0, min(100.0, score))
    color = _composite_color(score)
    pct = round(score, 1)
    return (
        f'<tr class="axis-row">'
        f'<td style="width:30%">{_esc(label)}</td>'
        f'<td><div class="bar-track"><div class="bar-fill" style="width:{pct}%;background:{color}"></div></div></td>'
        f'<td style="width:5rem;text-align:right;font-variant-numeric:tabular-nums">{pct:.0f}</td>'
        f'</tr>'
    )


def render_sparkline(values: list[float], width: int = 120, height: int = 24) -> str:
    """Inline SVG sparkline of composite scores over time.

    Newest-first input is reversed so left-to-right reads oldest-to-newest.
    """
    if not values:
        return ""
    series = list(reversed(values))
    if len(series) == 1:
        # Degenerate one-point case: draw a horizontal dot at mid-height.
        x = width // 2
        y = height // 2
        return (
            f'<svg class="spark" width="{width}" height="{height}" viewBox="0 0 {width} {height}">'
            f'<circle cx="{x}" cy="{y}" r="2" fill="{_composite_color(series[0])}"/>'
            f'</svg>'
        )
    hi = 100.0
    lo = 0.0
    n = len(series)
    points = []
    for i, v in enumerate(series):
        x = (i / (n - 1)) * width
        y = height - ((v - lo) / (hi - lo)) * height
        points.append(f"{x:.1f},{y:.1f}")
    polyline = " ".join(points)
    last_color = _composite_color(series[-1])
    return (
        f'<svg class="spark" width="{width}" height="{height}" viewBox="0 0 {width} {height}">'
        f'<polyline fill="none" stroke="{last_color}" stroke-width="1.5" points="{polyline}"/>'
        f'</svg>'
    )


def render_repo_page(
    output_dir: Path,
    slug: str,
    label: str,
    repo,  # RepoConfig
    history: list[dict],
    health: dict,
) -> None:
    """Write <output>/<slug>/index.html + feed.xml for one repo."""
    target_dir = output_dir / slug
    target_dir.mkdir(parents=True, exist_ok=True)

    if not history:
        return

    latest = history[0]
    radar_axes = (health.get("radar") or {}).get("axes") or {}

    axis_rows: list[str] = []
    # Stable axis order — keep core axes first.
    for key in ("complexity", "dead_code", "cycles", "coupling", "test_gap", "churn_surface"):
        score = latest["axes"].get(key)
        if score is None:
            score = float((radar_axes.get(key) or {}).get("score", 0.0) or 0.0)
        axis_rows.append(render_axis_bar(_AXIS_LABELS.get(key, key), float(score)))

    composites = [float(r["composite"]) for r in history]
    spark = render_sparkline(composites, width=180, height=32)

    history_rows: list[str] = []
    for r in history[:12]:
        sha_short = (r.get("sha") or "")[:7]
        history_rows.append(
            f'<tr>'
            f'<td><span class="muted">{_esc(_format_iso_date(r.get("timestamp", "")))}</span></td>'
            f'<td><code>{_esc(sha_short)}</code></td>'
            f'<td><span class="grade" style="background:{_grade_color(r.get("grade","?"))};color:#fff">{_esc(r.get("grade","?"))}</span></td>'
            f'<td style="font-variant-numeric:tabular-nums">{float(r.get("composite",0)):.1f}</td>'
            f'</tr>'
        )

    blurb_html = f'<p class="muted">{_esc(repo.blurb)}</p>' if repo.blurb else ""

    repo_url = getattr(repo, "url", "")
    grade_color = _grade_color(latest.get("grade", "?"))

    page = f"""<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{_esc(label)} — jCodeMunch Health Report</title>
<link rel="alternate" type="application/rss+xml" title="{_esc(label)} health feed" href="feed.xml">
<style>{_BASE_CSS}</style>
</head>
<body>
<p><a href="../index.html">← all reports</a></p>
<h1>{_esc(label)}</h1>
{blurb_html}
<p class="muted"><a href="{_esc(repo_url)}" target="_blank" rel="noopener">{_esc(repo_url)}</a> · <a href="feed.xml">RSS</a></p>

<div class="card" style="display:flex;align-items:center;gap:1.5rem;flex-wrap:wrap">
  <div>
    <div class="muted" style="font-size:0.85rem">Composite</div>
    <div style="font-size:2.25rem;font-weight:700;color:{_composite_color(latest['composite'])}">
      {float(latest['composite']):.1f}<span class="muted" style="font-size:1.25rem"> / 100</span>
    </div>
    <div><span class="grade" style="background:{grade_color};color:#fff;font-size:1rem">{_esc(latest.get('grade','?'))}</span></div>
  </div>
  <div style="flex:1 1 12rem;min-width:12rem">
    <div class="muted" style="font-size:0.85rem">Composite trend</div>
    {spark}
    <div class="muted" style="font-size:0.85rem">{len(history)} run(s) recorded</div>
  </div>
  <div style="flex:1 1 12rem;min-width:12rem">
    <div class="muted" style="font-size:0.85rem">Latest tree</div>
    <div><code>{_esc((latest.get('sha') or '')[:12])}</code></div>
    <div class="muted" style="font-size:0.85rem">{_esc(_format_iso_date(latest.get('timestamp','')))}</div>
  </div>
</div>

<h2>Six-axis radar</h2>
<div class="card">
  <table>
    <tbody>{''.join(axis_rows)}</tbody>
  </table>
</div>

<h2>History (latest 12 runs)</h2>
<div class="card">
  <table>
    <thead><tr><th>When</th><th>SHA</th><th>Grade</th><th style="text-align:right">Composite</th></tr></thead>
    <tbody>{''.join(history_rows)}</tbody>
  </table>
</div>

<p class="footer">Powered by <a href="https://github.com/jgravelle/jcodemunch-mcp">jcodemunch-mcp</a> ·
methodology in <a href="https://github.com/jgravelle/jcodemunch-mcp/blob/main/src/jcodemunch_mcp/tools/health_radar.py">health_radar.py</a> ·
heuristic, not coverage data.</p>
</body>
</html>
"""
    (target_dir / "index.html").write_text(page, encoding="utf-8")
    _write_repo_feed(target_dir, label, repo_url, history)


def _write_repo_feed(target_dir: Path, label: str, repo_url: str, history: list[dict]) -> None:
    """Per-repo RSS 2.0 feed of the last 24 runs."""
    items: list[str] = []
    for r in history[:24]:
        title = f"{r.get('grade','?')} ({float(r.get('composite',0)):.1f}) at {(r.get('sha') or '')[:7]}"
        desc = _sax.escape(r.get("summary", ""))
        guid = f"{label}::{r.get('sha','')}::{r.get('timestamp','')}"
        # RFC 822 pubDate; fall back to current time if parsing fails.
        try:
            dt = datetime.fromisoformat(r.get("timestamp", "").replace("Z", "+00:00"))
            pubdate = dt.strftime("%a, %d %b %Y %H:%M:%S +0000")
        except ValueError:
            pubdate = datetime.now(timezone.utc).strftime("%a, %d %b %Y %H:%M:%S +0000")
        items.append(
            f"<item>"
            f"<title>{_sax.escape(title)}</title>"
            f"<description>{desc}</description>"
            f"<guid isPermaLink=\"false\">{_sax.escape(guid)}</guid>"
            f"<pubDate>{pubdate}</pubDate>"
            f"</item>"
        )
    feed = (
        '<?xml version="1.0" encoding="UTF-8"?>'
        '<rss version="2.0">'
        '<channel>'
        f'<title>{_sax.escape(label)} — jCodeMunch Health</title>'
        f'<link>{_sax.escape(repo_url)}</link>'
        f'<description>Code-health radar runs for {_sax.escape(label)}</description>'
        f'<language>en-us</language>'
        + "".join(items) +
        '</channel>'
        '</rss>'
    )
    (target_dir / "feed.xml").write_text(feed, encoding="utf-8")


def render_index_page(output_dir: Path, summaries: list[dict]) -> None:
    """Top-level leaderboard, sorted by composite (desc)."""
    output_dir.mkdir(parents=True, exist_ok=True)
    ok_summaries = [s for s in summaries if s.get("status") == "ok"]
    failed = [s for s in summaries if s.get("status") != "ok"]
    ok_summaries.sort(key=lambda s: float(s.get("composite", 0)), reverse=True)

    tiles: list[str] = []
    for s in ok_summaries:
        slug = s["slug"]
        label = s.get("label", slug)
        composite = float(s.get("composite", 0))
        grade = s.get("grade", "?")
        history = []
        history_path = output_dir / slug / "history.json"
        if history_path.exists():
            try:
                history = json.loads(history_path.read_text(encoding="utf-8"))
            except (OSError, json.JSONDecodeError):
                history = []
        composites = [float(r["composite"]) for r in history]
        spark = render_sparkline(composites, width=120, height=24)
        # Phase 7: presence of runtime evidence is rare in OSS observatory
        # checkouts (most repos haven't had OTel traces ingested), so the
        # badge is a small aspirational nudge rather than a leaderboard
        # column. Hidden when False to keep the tile uncluttered.
        runtime_badge = (
            '<span class="runtime-badge" '
            'title="This repo has ingested runtime evidence (Phase 7)" '
            'style="font-size:0.7rem;padding:0.1rem 0.4rem;border-radius:3px;'
            'background:#1f6feb;color:#fff;margin-left:0.4rem">live</span>'
            if s.get("runtime_evidence") else ""
        )
        tiles.append(
            f'<a class="tile" href="{_esc(slug)}/index.html">'
            f'<div style="display:flex;align-items:center;justify-content:space-between;gap:0.5rem">'
            f'<h3>{_esc(label)}{runtime_badge}</h3>'
            f'<span class="grade" style="background:{_grade_color(grade)};color:#fff">{_esc(grade)}</span>'
            f'</div>'
            f'<div style="display:flex;align-items:center;gap:0.75rem;margin-top:0.5rem">'
            f'<span style="font-variant-numeric:tabular-nums;font-size:1.25rem;font-weight:600;color:{_composite_color(composite)}">{composite:.1f}</span>'
            f'<span class="muted">/ 100</span>'
            f'<span style="margin-left:auto">{spark}</span>'
            f'</div>'
            f'</a>'
        )

    failed_block = ""
    if failed:
        rows = []
        for s in failed:
            rows.append(f'<li>{_esc(s.get("label", s.get("slug","?")))} — <span class="muted">{_esc(s.get("status","unknown"))}</span></li>')
        failed_block = (
            f'<h2>Skipped</h2>'
            f'<div class="card"><ul>{"".join(rows)}</ul></div>'
        )

    page = f"""<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>jCodeMunch — OSS Health Observatory</title>
<link rel="alternate" type="application/rss+xml" title="Latest health updates" href="feed.xml">
<style>{_BASE_CSS}</style>
</head>
<body>
<h1>jCodeMunch — OSS Health Observatory</h1>
<p class="muted">Six-axis code-health radar over a curated list of OSS repos
(plus an optional seventh axis when runtime evidence is available — repos
with the <span style="font-size:0.75rem;padding:0.05rem 0.3rem;border-radius:3px;background:#1f6feb;color:#fff">live</span> badge have ingested OTel / SQL / stack traces).
Re-audited periodically. Grades are heuristic — read the
<a href="https://github.com/jgravelle/jcodemunch-mcp/blob/main/src/jcodemunch_mcp/tools/health_radar.py">methodology</a>
before using as anything more than a directional signal.
<a href="feed.xml">RSS</a></p>

<h2>Tracked repositories</h2>
<div class="repo-list">
{''.join(tiles)}
</div>

{failed_block}

<p class="footer">Powered by <a href="https://github.com/jgravelle/jcodemunch-mcp">jcodemunch-mcp</a>.
Each tile links to the repo's full radar + history. RSS feeds available per-repo.</p>
</body>
</html>
"""
    (output_dir / "index.html").write_text(page, encoding="utf-8")
    # Mirror as JSON for downstream tooling.  Top-level metadata makes the
    # artifact self-describing — verifiers can confirm which jcm version +
    # INDEX_VERSION produced the run from the file alone, no workflow log
    # cross-reference required.
    try:
        from .. import __version__ as _gen_version
    except ImportError:
        _gen_version = ""
    try:
        from ..storage.index_store import INDEX_VERSION as _gen_index_version
    except ImportError:
        _gen_index_version = 0
    payload = {
        "generator_version": _gen_version,
        "index_version": _gen_index_version,
        "built_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "summaries": ok_summaries,
        "skipped": failed,
    }
    (output_dir / "index.json").write_text(
        json.dumps(payload, indent=2) + "\n",
        encoding="utf-8",
    )


def render_index_feed(output_dir: Path, summaries: list[dict]) -> None:
    """Cross-repo RSS feed: one item per ok-summary's latest run."""
    items: list[str] = []
    for s in summaries:
        if s.get("status") != "ok":
            continue
        slug = s["slug"]
        label = s.get("label", slug)
        title = f"{label}: {s.get('grade','?')} ({float(s.get('composite',0)):.1f})"
        guid = f"{slug}::{s.get('sha','')}"
        items.append(
            f"<item>"
            f"<title>{_sax.escape(title)}</title>"
            f"<link>./{_sax.escape(slug)}/index.html</link>"
            f"<guid isPermaLink=\"false\">{_sax.escape(guid)}</guid>"
            f"</item>"
        )
    feed = (
        '<?xml version="1.0" encoding="UTF-8"?>'
        '<rss version="2.0">'
        '<channel>'
        '<title>jCodeMunch — OSS Health Observatory</title>'
        '<link>./index.html</link>'
        '<description>Latest code-health audits across the tracked OSS list.</description>'
        '<language>en-us</language>'
        + "".join(items) +
        '</channel>'
        '</rss>'
    )
    (output_dir / "feed.xml").write_text(feed, encoding="utf-8")
