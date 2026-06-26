"""Render a sticky PR comment from baseline + PR radar JSON files.

Reads two radar payloads (the `radar` sub-field of `jcodemunch-mcp
health` output), computes the diff via the same pure helper the MCP
tool exposes, and prints a markdown PR comment to stdout.

Stable HTML comment marker on the first line lets the calling shell
script find/edit an existing comment instead of spamming the PR.
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Iterable

# This script runs in the Actions runner where jcodemunch-mcp has been
# pip-installed, so the import works directly.
from jcodemunch_mcp.tools.health_radar import diff_radar


_MARKER = "<!-- jcm-health-radar -->"
_GRADE_EMOJI = {"A": "🟢", "B": "🟢", "C": "🟡", "D": "🟠", "F": "🔴"}


def _arrow(delta) -> str:
    if delta is None:
        return "—"
    if delta >= 3.0:
        return "↑"
    if delta <= -3.0:
        return "↓"
    return "·"


def _signed(v) -> str:
    if v is None:
        return "—"
    if v > 0:
        return f"+{v:.1f}"
    return f"{v:.1f}"


def _radar_table(diff: dict) -> str:
    lines = [
        "| Axis | Baseline | PR | Δ |",
        "|---|---:|---:|---:|",
    ]
    for axis, d in sorted(diff["axis_deltas"].items()):
        b = d.get("from")
        c = d.get("to")
        delta = d.get("delta")
        if b is None and c is None:
            continue
        b_s = "—" if b is None else f"{b:.0f}"
        c_s = "—" if c is None else f"{c:.0f}"
        delta_s = _signed(delta)
        emphasis = "**" if delta is not None and abs(delta) >= 3.0 else ""
        lines.append(
            f"| `{axis}` | {b_s} | {c_s} | {emphasis}{delta_s}{emphasis} {_arrow(delta)} |"
        )
    return "\n".join(lines)


def _regression_detail(diff: dict, kind: str) -> str:
    """Render the bulleted regression / improvement axes with raw values."""
    items = diff.get(kind, [])
    if not items:
        return ""
    lines = [f"### {'Regressions' if kind == 'regressions' else 'Improvements'}"]
    for axis in items:
        d = diff["axis_deltas"][axis]
        raw_from = d.get("raw_from")
        raw_to = d.get("raw_to")
        if raw_from is not None and raw_to is not None:
            lines.append(f"- `{axis}`: raw {raw_from} → {raw_to}")
        else:
            lines.append(f"- `{axis}`: score {d.get('from')} → {d.get('to')}")
    return "\n".join(lines)


def _verdict_emoji(diff: dict) -> str:
    if diff.get("regressions") and not diff.get("improvements"):
        return "🔴"
    if diff.get("improvements") and not diff.get("regressions"):
        return "🟢"
    if diff.get("regressions") and diff.get("improvements"):
        return "🟡"
    return "⚪"


def render(baseline: dict, current: dict, *, version: str = "") -> str:
    diff = diff_radar(baseline, current)

    base_grade = baseline.get("grade", "?")
    cur_grade = current.get("grade", "?")
    grade_line = (
        f"{_GRADE_EMOJI.get(cur_grade, '')} **Composite:** "
        f"{base_grade} → {cur_grade} ({_signed(diff['composite_delta'])} pts)"
    )

    blocks: list[str] = [
        _MARKER,
        "## jCodeMunch Health Radar",
        "",
        grade_line,
        f"{_verdict_emoji(diff)} **Verdict:** {diff['verdict']}",
        "",
        _radar_table(diff),
    ]

    regressions = _regression_detail(diff, "regressions")
    if regressions:
        blocks.append("")
        blocks.append(regressions)

    improvements = _regression_detail(diff, "improvements")
    if improvements:
        blocks.append("")
        blocks.append(improvements)

    blocks.append("")
    blocks.append(
        "_[Methodology](https://github.com/jgravelle/jcodemunch-mcp/blob/main/src/jcodemunch_mcp/tools/health_radar.py) "
        f"· Posted by `jcodemunch-mcp{(' v' + version) if version else ''}`. "
        "Re-runs on every push to this PR; this comment is sticky._"
    )
    return "\n".join(blocks).rstrip() + "\n"


def _load_radar(path: Path) -> dict:
    """Load a radar payload from JSON. Accepts either the `radar` sub-dict
    or the full get_repo_health response (we'll unwrap)."""
    payload = json.loads(path.read_text(encoding="utf-8"))
    if isinstance(payload, dict) and "axes" in payload:
        return payload
    if isinstance(payload, dict) and "radar" in payload:
        return payload["radar"]
    raise ValueError(
        f"{path} doesn't look like a radar payload (no `axes` field, no `radar` sub-field)"
    )


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(
        description="Render a sticky PR comment from baseline + PR radar JSON.",
    )
    parser.add_argument("--baseline", type=Path, required=True,
        help="Path to baseline radar JSON (e.g. base branch).")
    parser.add_argument("--current", type=Path, required=True,
        help="Path to current radar JSON (e.g. PR branch).")
    parser.add_argument("--version", default="",
        help="jcodemunch-mcp version string for the footer.")
    args = parser.parse_args(argv)

    try:
        baseline = _load_radar(args.baseline)
        current = _load_radar(args.current)
    except (OSError, json.JSONDecodeError, ValueError) as e:
        print(f"error: {e}", file=sys.stderr)
        return 2

    sys.stdout.write(render(baseline, current, version=args.version))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
