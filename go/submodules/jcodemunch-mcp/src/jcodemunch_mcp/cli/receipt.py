"""``jcodemunch-mcp receipt`` — token-economy ledger.

Parses ``~/.claude/projects/**/*.jsonl`` transcripts, extracts every
``mcp__jcodemunch__*`` tool call + its result, applies per-tool savings
multipliers calibrated against the published RAG benchmarks, and prints
an honest dollar-denominated ROI ledger.

The savings model is **modeled, not measured** — token-savings is
inherently counterfactual (we can't observe what naive Read+Grep would
have cost without running it). The methodology is auditable via
``--explain``; raw per-call data is exportable via ``--export``.
"""

from __future__ import annotations

import argparse
import collections
import csv
import datetime as _dt
import io
import json
import sys
from pathlib import Path
from typing import Iterable, Optional

# Per-tool savings multipliers: jcodemunch result tokens × multiplier ≈
# what naive Read+Grep would have spent answering the same query.
# Calibrated to be *conservative* against published RAG benchmarks
# (which show 30–56× reductions for retrieval-style queries on Express,
# FastAPI, Gin). Underestimating savings keeps credibility — a number
# that's plausibly low is more useful than one that's optimistic.
_TOOL_MULTIPLIERS: dict[str, float] = {
    # Pure retrieval — narrow query against indexed corpus.
    "search_symbols": 20.0,
    "search_text": 12.0,
    "search_columns": 15.0,
    "search_ast": 18.0,
    "get_ranked_context": 18.0,
    "winnow_symbols": 15.0,
    # Targeted symbol/file fetch — surgical vs whole-file Read.
    "get_symbol_source": 8.0,
    "get_context_bundle": 10.0,
    "get_file_outline": 6.0,
    "get_file_content": 2.0,  # nearly 1:1, only saves on filtering
    # Repo structure / orientation.
    "get_repo_outline": 8.0,
    "get_file_tree": 4.0,
    "get_project_intel": 12.0,
    "get_session_context": 6.0,
    "get_session_snapshot": 6.0,
    # Graph queries — the structurally hardest things to do with grep.
    "find_importers": 25.0,
    "find_references": 25.0,
    "check_references": 15.0,
    "get_call_hierarchy": 30.0,
    "get_dependency_graph": 25.0,
    "get_dependency_cycles": 25.0,
    "get_blast_radius": 35.0,
    "get_class_hierarchy": 20.0,
    "get_layer_violations": 20.0,
    "get_extraction_candidates": 25.0,
    "get_signal_chains": 30.0,
    "get_cross_repo_map": 25.0,
    "get_related_symbols": 18.0,
    # Risk / health / quality — composite metrics that have no naive
    # equivalent (you'd have to write the analysis yourself).
    "get_pr_risk_profile": 40.0,
    "get_repo_health": 35.0,
    "get_hotspots": 25.0,
    "get_symbol_complexity": 12.0,
    "get_churn_rate": 6.0,
    "get_symbol_provenance": 15.0,
    "get_untested_symbols": 30.0,
    "find_dead_code": 35.0,
    "get_dead_code_v2": 35.0,
    "get_tectonic_map": 30.0,
    "get_coupling_metrics": 20.0,
    "get_symbol_importance": 15.0,
    # Refactoring / maintenance.
    "plan_refactoring": 25.0,
    "check_rename_safe": 15.0,
    "get_symbol_diff": 12.0,
    "get_changed_symbols": 12.0,
    "audit_agent_config": 8.0,
    # Indexing / repo management.
    "resolve_repo": 3.0,
    "list_repos": 2.0,
    "index_folder": 2.0,
    "index_repo": 2.0,
    "index_file": 2.0,
}

# Default multiplier for tools not in the table above. Conservative
# middle-of-the-road estimate.
_DEFAULT_MULTIPLIER = 8.0

# Model prices in USD per million input tokens. Cache-read pricing is
# typically 10% of normal input pricing for Anthropic models, but we use
# normal input pricing here because savings are computed against a
# counterfactual (naive Read+Grep would have been *fresh* input, not
# cached). Opus is the default — most jcodemunch users are running an
# Opus-grade model where savings actually move a budget needle.
_MODEL_PRICES_USD_PER_MTOK: dict[str, float] = {
    "sonnet": 3.0,    # Claude Sonnet 4.x
    "opus":   15.0,   # Claude Opus 4.x
    "haiku":  0.80,   # Claude Haiku 4.x
}

_DEFAULT_MODEL = "opus"

# Approximate bytes-per-token used to convert tool_result content
# byte-length into a token estimate. Same heuristic the rest of the
# package uses (see _BYTES_PER_TOKEN in storage/token_tracker.py).
_BYTES_PER_TOKEN = 4


def _projects_root() -> Path:
    return Path.home() / ".claude" / "projects"


def _result_byte_length(content) -> int:
    """Return the byte length of a tool_result `content` field.

    Claude Code stores tool_result.content as either a string or a list
    of content blocks ({type: 'text', text: '...'}). Sum text lengths;
    other block types contribute nothing to the token estimate.
    """
    if content is None:
        return 0
    if isinstance(content, str):
        return len(content.encode("utf-8", errors="replace"))
    if isinstance(content, list):
        total = 0
        for block in content:
            if isinstance(block, dict) and block.get("type") == "text":
                t = block.get("text") or ""
                total += len(t.encode("utf-8", errors="replace"))
        return total
    return 0


def _parse_iso(ts: str) -> Optional[_dt.datetime]:
    if not ts:
        return None
    try:
        # Claude Code timestamps end in 'Z'.
        return _dt.datetime.fromisoformat(ts.replace("Z", "+00:00"))
    except ValueError:
        return None


def iter_calls(
    projects_root: Path,
    *,
    since: Optional[_dt.datetime] = None,
) -> Iterable[dict]:
    """Yield {tool, result_tokens, timestamp, project, session} per jcodemunch call.

    Walks the entire ~/.claude/projects/ tree. For each tool_use block
    naming an mcp__jcodemunch__* tool, finds the matching tool_result
    (by tool_use_id) in subsequent user events within the same session
    file, and yields one entry per resolved pair.
    """
    if not projects_root.exists():
        return

    for jsonl in sorted(projects_root.rglob("*.jsonl")):
        try:
            yield from _iter_calls_in_file(jsonl, since=since)
        except OSError:
            continue


def _iter_calls_in_file(
    jsonl: Path,
    *,
    since: Optional[_dt.datetime],
) -> Iterable[dict]:
    """Walk one transcript file once; pair tool_use → tool_result by id."""
    pending: dict[str, dict] = {}  # tool_use_id → call metadata

    try:
        with open(jsonl, encoding="utf-8", errors="ignore") as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    ev = json.loads(line)
                except json.JSONDecodeError:
                    continue
                if not isinstance(ev, dict):
                    continue
                msg = ev.get("message")
                if not isinstance(msg, dict):
                    continue
                content = msg.get("content")
                if not isinstance(content, list):
                    continue

                ts_raw = ev.get("timestamp", "")
                ts = _parse_iso(ts_raw)
                if since and ts and ts < since:
                    # Per-event since filter — but we still walk the whole
                    # file because session files aren't strictly ordered
                    # by event timestamp.
                    pass

                for block in content:
                    if not isinstance(block, dict):
                        continue
                    btype = block.get("type")

                    if btype == "tool_use":
                        name = block.get("name") or ""
                        if not name.startswith("mcp__jcodemunch"):
                            continue
                        tu_id = block.get("id") or ""
                        if not tu_id:
                            continue
                        # Strip the mcp__jcodemunch__ prefix to get the
                        # bare tool name (search_symbols, etc.).
                        bare = name.split("__")[-1] if "__" in name else name
                        pending[tu_id] = {
                            "tool": bare,
                            "timestamp": ts_raw,
                            "_ts_parsed": ts,
                        }

                    elif btype == "tool_result":
                        tu_id = block.get("tool_use_id") or ""
                        if not tu_id or tu_id not in pending:
                            continue
                        meta = pending.pop(tu_id)
                        if since and meta["_ts_parsed"] and meta["_ts_parsed"] < since:
                            continue
                        result_bytes = _result_byte_length(block.get("content"))
                        result_tokens = max(1, result_bytes // _BYTES_PER_TOKEN)
                        yield {
                            "tool": meta["tool"],
                            "timestamp": meta["timestamp"],
                            "result_tokens": result_tokens,
                            "result_bytes": result_bytes,
                            "session_file": str(jsonl),
                        }
    except OSError:
        return


def aggregate(calls: Iterable[dict]) -> dict:
    """Aggregate per-tool savings from a stream of call records."""
    per_tool: dict[str, dict] = collections.defaultdict(
        lambda: {"calls": 0, "actual_tokens": 0, "baseline_tokens": 0, "savings_tokens": 0}
    )
    total_calls = 0
    for call in calls:
        tool = call["tool"]
        actual = call["result_tokens"]
        mult = _TOOL_MULTIPLIERS.get(tool, _DEFAULT_MULTIPLIER)
        baseline = int(actual * mult)
        savings = baseline - actual

        bucket = per_tool[tool]
        bucket["calls"] += 1
        bucket["actual_tokens"] += actual
        bucket["baseline_tokens"] += baseline
        bucket["savings_tokens"] += savings
        total_calls += 1

    totals = {
        "calls": total_calls,
        "actual_tokens": sum(b["actual_tokens"] for b in per_tool.values()),
        "baseline_tokens": sum(b["baseline_tokens"] for b in per_tool.values()),
        "savings_tokens": sum(b["savings_tokens"] for b in per_tool.values()),
    }
    return {"totals": totals, "per_tool": dict(per_tool)}


def dollar_savings(savings_tokens: int, model: str) -> float:
    rate = _MODEL_PRICES_USD_PER_MTOK.get(model.lower())
    if rate is None:
        return 0.0
    return (savings_tokens / 1_000_000.0) * rate


def render_text(agg: dict, *, days: int, model: str, primary_only: bool = False) -> str:
    """Human-readable ledger output."""
    out = io.StringIO()
    totals = agg["totals"]
    per_tool = agg["per_tool"]
    out.write(f"jCodeMunch token-economy ledger — last {days} days\n")
    out.write("=" * 56 + "\n\n")

    if totals["calls"] == 0:
        out.write("No jcodemunch tool calls found in ~/.claude/projects/.\n")
        out.write("If you've been using jcodemunch, check that Claude Code is\n")
        out.write("writing transcripts to that directory (default behaviour).\n")
        return out.getvalue()

    out.write(f"  Tool calls:                    {totals['calls']:>12,}\n")
    out.write(f"  Tokens delivered (actual):     {totals['actual_tokens']:>12,}\n")
    out.write(f"  Tokens you would have spent:   {totals['baseline_tokens']:>12,}\n")
    out.write(f"                                 {'-' * 12}\n")
    out.write(f"  Net savings:                   {totals['savings_tokens']:>12,} tokens\n\n")

    rate = _MODEL_PRICES_USD_PER_MTOK.get(model.lower(), _MODEL_PRICES_USD_PER_MTOK[_DEFAULT_MODEL])
    primary_dollars = dollar_savings(totals["savings_tokens"], model)
    out.write(f"  Saved at {model.title()} pricing (${rate:.2f}/MTok input):  ${primary_dollars:,.2f}\n")

    if not primary_only:
        for other in ("sonnet", "opus", "haiku"):
            if other == model.lower():
                continue
            other_rate = _MODEL_PRICES_USD_PER_MTOK[other]
            other_dollars = dollar_savings(totals["savings_tokens"], other)
            out.write(f"     ... at {other.title()} pricing (${other_rate:.2f}/MTok):                 ${other_dollars:,.2f}\n")
    out.write("\n")

    if per_tool:
        out.write("  Top tools by savings:\n")
        ranked = sorted(
            per_tool.items(),
            key=lambda kv: kv[1]["savings_tokens"],
            reverse=True,
        )[:10]
        out.write(f"    {'tool':<28} {'calls':>8} {'savings (tokens)':>20}\n")
        for name, b in ranked:
            out.write(f"    {name:<28} {b['calls']:>8,} {b['savings_tokens']:>20,}\n")
        out.write("\n")

    out.write("  Methodology: per-tool savings multipliers calibrated against\n")
    out.write("  published RAG benchmarks (Express/FastAPI/Gin). Run with --explain\n")
    out.write("  to see the full multiplier table; --export csv|json for raw data.\n")

    return out.getvalue()


def render_explain() -> str:
    """Per-tool multiplier table + methodology notes."""
    out = io.StringIO()
    out.write("jcodemunch receipt — savings model methodology\n")
    out.write("=" * 56 + "\n\n")
    out.write("Per-tool savings multipliers. For each call:\n")
    out.write("  baseline_tokens = actual_tokens × multiplier\n")
    out.write("  savings_tokens  = baseline_tokens − actual_tokens\n\n")
    out.write("Calibrated against published RAG benchmarks\n")
    out.write("(benchmarks/rag_baseline_results.md) which show 30–56×\n")
    out.write("retrieval savings on Express/FastAPI/Gin. Multipliers below\n")
    out.write("are deliberately conservative — underestimating savings keeps\n")
    out.write("the dollar number defensible.\n\n")
    out.write(f"Default multiplier (unlisted tools): {_DEFAULT_MULTIPLIER}×\n\n")

    rows = sorted(_TOOL_MULTIPLIERS.items(), key=lambda kv: kv[1], reverse=True)
    out.write(f"  {'tool':<32} {'multiplier':>12}\n")
    for tool, mult in rows:
        out.write(f"  {tool:<32} {mult:>11.1f}×\n")
    out.write("\nTo override a tool's multiplier, edit cli/receipt.py and\n")
    out.write("send a PR with your reasoning. The numbers should reflect\n")
    out.write("realistic naive-tool-call counterfactuals, not optimism.\n")
    return out.getvalue()


def render_csv(agg: dict) -> str:
    out = io.StringIO()
    w = csv.writer(out)
    w.writerow(["tool", "calls", "actual_tokens", "baseline_tokens", "savings_tokens"])
    for tool, b in sorted(agg["per_tool"].items()):
        w.writerow([tool, b["calls"], b["actual_tokens"], b["baseline_tokens"], b["savings_tokens"]])
    return out.getvalue()


def render_json(agg: dict, *, model: str) -> str:
    payload = {
        "totals": agg["totals"],
        "per_tool": agg["per_tool"],
        "model": model,
        "savings_usd": dollar_savings(agg["totals"]["savings_tokens"], model),
    }
    return json.dumps(payload, indent=2)


def main(argv: Optional[list[str]] = None) -> int:
    parser = argparse.ArgumentParser(
        description="jcodemunch-mcp receipt — token-economy ledger from Claude Code transcripts.",
    )
    parser.add_argument(
        "--days",
        type=int,
        default=30,
        help="Window size in days (default 30; use 0 for all-time).",
    )
    parser.add_argument(
        "--model",
        choices=sorted(_MODEL_PRICES_USD_PER_MTOK.keys()),
        default=_DEFAULT_MODEL,
        help="Model rate to apply for the dollar conversion (default opus).",
    )
    parser.add_argument(
        "--export",
        metavar="FILE.csv|FILE.json",
        help="Write raw per-tool data to a file instead of the human report.",
    )
    parser.add_argument(
        "--explain",
        action="store_true",
        help="Print the per-tool savings multiplier table + methodology, then exit.",
    )
    parser.add_argument(
        "--projects-root",
        type=Path,
        default=None,
        help="Override Claude Code projects directory (default ~/.claude/projects).",
    )
    args = parser.parse_args(argv)

    if args.explain:
        sys.stdout.write(render_explain())
        return 0

    root = args.projects_root or _projects_root()
    since = None
    if args.days > 0:
        since = _dt.datetime.now(_dt.timezone.utc) - _dt.timedelta(days=args.days)

    agg = aggregate(iter_calls(root, since=since))

    if args.export:
        target = Path(args.export)
        ext = target.suffix.lower()
        if ext == ".json":
            target.write_text(render_json(agg, model=args.model), encoding="utf-8")
        elif ext == ".csv":
            target.write_text(render_csv(agg), encoding="utf-8")
        else:
            print(
                f"--export needs a .csv or .json filename, got {target}",
                file=sys.stderr,
            )
            return 2
        print(f"wrote {target}")
        return 0

    sys.stdout.write(render_text(agg, days=args.days, model=args.model))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
