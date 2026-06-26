"""audit_agent_config — find wasted tokens in agent configuration files.

Scans CLAUDE.md, .cursorrules, copilot-instructions.md, and other agent
config files for token cost, stale symbol/file references, redundancy,
and bloat.  Cross-references against the jcodemunch index to catch dead
paths and renamed symbols that no other linter can detect.
"""

from __future__ import annotations

import logging
import os
import re
from pathlib import Path
from typing import Any, Optional

from ..storage import IndexStore

logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Token estimation
# ---------------------------------------------------------------------------

def _estimate_tokens(text: str) -> int:
    """Estimate token count.  Uses tiktoken if available, else word heuristic."""
    try:
        import tiktoken
        enc = tiktoken.get_encoding("cl100k_base")
        return len(enc.encode(text))
    except Exception:
        # Rough heuristic: ~0.75 tokens per word for English prose/code mix
        return max(1, int(len(text.split()) * 1.3))


# ---------------------------------------------------------------------------
# Config file discovery
# ---------------------------------------------------------------------------

# (relative_to_home, relative_to_project, description)
_GLOBAL_CONFIGS: list[tuple[str, str]] = [
    (os.path.join(".claude", "CLAUDE.md"), "global CLAUDE.md"),
    (os.path.join(".claude", "settings.json"), "Claude Code settings"),
]

_PROJECT_CONFIGS: list[tuple[str, str]] = [
    ("CLAUDE.md", "project CLAUDE.md"),
    (".cursorrules", "Cursor rules"),
    (os.path.join(".cursor", "rules"), "Cursor rules (new)"),
    (os.path.join(".github", "copilot-instructions.md"), "GitHub Copilot instructions"),
    (".windsurfrules", "Windsurf rules"),
    (os.path.join(".antigravity", "rules"), "Google Antigravity rules"),
    (os.path.join(".continue", "config.json"), "Continue config"),
]


def _discover_files(project_path: Optional[str] = None) -> list[dict[str, Any]]:
    """Find all agent config files and return metadata."""
    home = Path.home()
    project = Path(project_path) if project_path else None
    found: list[dict[str, Any]] = []

    for rel, desc in _GLOBAL_CONFIGS:
        p = home / rel
        if p.is_file():
            try:
                text = p.read_text(encoding="utf-8", errors="replace")
                found.append({
                    "path": str(p),
                    "description": desc,
                    "scope": "global",
                    "content": text,
                    "tokens": _estimate_tokens(text),
                    "lines": text.count("\n") + 1,
                })
            except OSError:
                pass

    if project:
        for rel, desc in _PROJECT_CONFIGS:
            p = project / rel
            if p.is_file():
                try:
                    text = p.read_text(encoding="utf-8", errors="replace")
                    found.append({
                        "path": str(p),
                        "description": desc,
                        "scope": "project",
                        "content": text,
                        "tokens": _estimate_tokens(text),
                        "lines": text.count("\n") + 1,
                    })
                except OSError:
                    pass

    return found


# ---------------------------------------------------------------------------
# Finding extractors
# ---------------------------------------------------------------------------

# Matches things that look like symbol names: camelCase, snake_case, PascalCase
# in contexts like "use X", "call X", "function X", backtick-quoted `X`
_BACKTICK_REF = re.compile(r"`([a-zA-Z_]\w{2,}(?:\.\w+)*)`")
_SYMBOL_LIKE = re.compile(
    r"(?:^|\s)(?:function|method|class|use|call|check|run|invoke|symbol)\s+"
    r"[`'\"]?([a-zA-Z_]\w{2,})[`'\"]?",
    re.IGNORECASE | re.MULTILINE,
)

# Matches file paths: src/foo/bar.py, ./utils/helpers.js, etc.
_FILE_PATH_REF = re.compile(
    r"(?:^|\s|[`'\"])((?:\./|src/|lib/|app/|test|pkg/|internal/|cmd/)"
    r"[a-zA-Z0-9_./-]+\.[a-zA-Z]{1,10})(?:\s|[`'\"]|$|:|\))",
    re.MULTILINE,
)


def _extract_symbol_refs(text: str) -> list[str]:
    """Extract potential symbol name references from config text."""
    refs: set[str] = set()
    for m in _BACKTICK_REF.finditer(text):
        candidate = m.group(1)
        # Filter out common non-symbol backtick content
        if not candidate.startswith(("http", "pip ", "npm ", "git ", "--")):
            refs.add(candidate)
    for m in _SYMBOL_LIKE.finditer(text):
        refs.add(m.group(1))
    return sorted(refs)


def _extract_file_refs(text: str) -> list[str]:
    """Extract potential file path references from config text."""
    refs: set[str] = set()
    for m in _FILE_PATH_REF.finditer(text):
        refs.add(m.group(1))
    return sorted(refs)


def _find_line(text: str, needle: str) -> Optional[int]:
    """Find the 1-based line number of needle in text."""
    for i, line in enumerate(text.splitlines(), 1):
        if needle in line:
            return i
    return None


# ---------------------------------------------------------------------------
# Checkers
# ---------------------------------------------------------------------------

def _check_stale_symbols(
    file_info: dict, index, all_symbol_names: set[str],
) -> list[dict[str, Any]]:
    """Find references to symbols that don't exist in the index."""
    findings: list[dict[str, Any]] = []
    refs = _extract_symbol_refs(file_info["content"])
    for ref in refs:
        # Check against symbol names (plain name, not full ID)
        parts = ref.split(".")
        name = parts[-1]  # Use the last segment (e.g., "bar" from "foo.bar")
        if name not in all_symbol_names and len(name) > 3:
            # Try fuzzy match for suggestions
            suggestion = _fuzzy_suggest(name, all_symbol_names)
            msg = f"References symbol '{ref}' which was not found in the index"
            if suggestion:
                msg += f". Did you mean '{suggestion}'?"
            findings.append({
                "file": file_info["path"],
                "severity": "warning",
                "category": "stale_reference",
                "message": msg,
                "line": _find_line(file_info["content"], ref),
            })
    return findings


def _check_dead_paths(
    file_info: dict, source_files: set[str], source_root: str,
) -> list[dict[str, Any]]:
    """Find references to files that don't exist in the index."""
    findings: list[dict[str, Any]] = []
    refs = _extract_file_refs(file_info["content"])
    for ref in refs:
        # Normalize: strip leading ./
        normalized = ref.lstrip("./")
        # Check if any source file ends with this path
        matched = any(sf.endswith(normalized) or sf == normalized for sf in source_files)
        if not matched:
            # Also check absolute path from source_root
            if source_root:
                abs_path = os.path.join(source_root, normalized)
                if os.path.exists(abs_path):
                    continue
            findings.append({
                "file": file_info["path"],
                "severity": "warning",
                "category": "dead_path",
                "message": f"References '{ref}' which does not exist in the indexed file tree",
                "line": _find_line(file_info["content"], ref),
            })
    return findings


def _check_redundancy(files: list[dict]) -> list[dict[str, Any]]:
    """Find duplicate content between global and project configs."""
    findings: list[dict[str, Any]] = []
    global_files = [f for f in files if f["scope"] == "global"]
    project_files = [f for f in files if f["scope"] == "project"]

    for gf in global_files:
        g_lines = [
            ln.strip() for ln in gf["content"].splitlines()
            if ln.strip() and not ln.strip().startswith("#")
        ]
        for pf in project_files:
            p_lines = [
                ln.strip() for ln in pf["content"].splitlines()
                if ln.strip() and not ln.strip().startswith("#")
            ]
            # Find runs of 3+ consecutive matching lines
            duplicates = _find_duplicate_runs(g_lines, p_lines, min_run=3)
            for dup in duplicates:
                findings.append({
                    "file": pf["path"],
                    "severity": "info",
                    "category": "redundancy",
                    "message": (
                        f"{dup['count']} lines duplicate content already present "
                        f"in {gf['path']} (starting with: '{dup['preview']}')"
                    ),
                    "line": None,
                })
    return findings


def _find_duplicate_runs(
    a_lines: list[str], b_lines: list[str], min_run: int = 3,
) -> list[dict]:
    """Find runs of min_run+ consecutive matching lines between two line lists."""
    if not a_lines or not b_lines:
        return []
    a_set = set(a_lines)
    runs: list[dict] = []
    current_run = 0
    first_line = ""
    for line in b_lines:
        if line in a_set:
            if current_run == 0:
                first_line = line
            current_run += 1
        else:
            if current_run >= min_run:
                preview = first_line[:80] + ("..." if len(first_line) > 80 else "")
                runs.append({"count": current_run, "preview": preview})
            current_run = 0
    if current_run >= min_run:
        preview = first_line[:80] + ("..." if len(first_line) > 80 else "")
        runs.append({"count": current_run, "preview": preview})
    return runs


def _check_bloat(file_info: dict) -> list[dict[str, Any]]:
    """Detect common bloat patterns."""
    findings: list[dict[str, Any]] = []
    content = file_info["content"]
    tokens = file_info["tokens"]

    # Large file warning
    if tokens > 2000:
        findings.append({
            "file": file_info["path"],
            "severity": "warning",
            "category": "bloat",
            "message": (
                f"This file is {tokens:,} tokens. Agent config files over 2,000 tokens "
                f"add significant per-turn cost. Consider trimming verbose examples or "
                f"moving reference material to external docs."
            ),
            "line": None,
        })

    # Excessive code blocks (examples that bloat context)
    code_blocks = re.findall(r"```[\s\S]*?```", content)
    if code_blocks:
        block_chars = sum(len(b) for b in code_blocks)
        block_pct = block_chars / max(len(content), 1)
        if block_pct > 0.5 and len(code_blocks) > 2:
            findings.append({
                "file": file_info["path"],
                "severity": "info",
                "category": "bloat",
                "message": (
                    f"{len(code_blocks)} code blocks make up {block_pct:.0%} of the file. "
                    f"Consider moving examples to a separate reference doc and linking to it."
                ),
                "line": None,
            })

    return findings


def _check_scope_leaks(file_info: dict) -> list[dict[str, Any]]:
    """Detect project-specific rules that leaked into global config."""
    if file_info["scope"] != "global":
        return []
    findings: list[dict[str, Any]] = []
    content = file_info["content"]

    # Patterns that suggest project-specific rules
    project_signals = [
        (r"(?:pytest|jest|mocha|vitest)\b.*(?:--\w+)", "test runner configuration"),
        (r"(?:django|flask|fastapi|express|rails)\b", "framework-specific rules"),
        (r"(?:src/|lib/|app/|pkg/)\w+/\w+\.\w+", "project-specific file paths"),
        (r"(?:table|migration|schema)\s+\w+", "database-specific references"),
    ]
    for pattern, desc in project_signals:
        matches = re.findall(pattern, content, re.IGNORECASE)
        if matches:
            sample = matches[0][:60]
            findings.append({
                "file": file_info["path"],
                "severity": "info",
                "category": "scope_leak",
                "message": (
                    f"Possible project-specific content in global config ({desc}: "
                    f"'{sample}'). Global rules cost tokens in every project."
                ),
                "line": _find_line(content, sample),
            })
            break  # One finding per file is enough
    return findings


def _fuzzy_suggest(name: str, candidates: set[str], max_dist: int = 3) -> Optional[str]:
    """Find the closest candidate by edit distance."""
    name_lower = name.lower()
    best: Optional[str] = None
    best_dist = max_dist + 1
    for c in candidates:
        if abs(len(c) - len(name)) > max_dist:
            continue
        d = _edit_distance(name_lower, c.lower(), max_dist)
        if d < best_dist:
            best_dist = d
            best = c
    return best if best_dist <= max_dist else None


def _edit_distance(a: str, b: str, max_d: int) -> int:
    """Levenshtein distance with early termination."""
    if abs(len(a) - len(b)) > max_d:
        return max_d + 1
    if a == b:
        return 0
    la, lb = len(a), len(b)
    prev = list(range(lb + 1))
    for i in range(1, la + 1):
        curr = [i] + [0] * lb
        for j in range(1, lb + 1):
            cost = 0 if a[i - 1] == b[j - 1] else 1
            curr[j] = min(curr[j - 1] + 1, prev[j] + 1, prev[j - 1] + cost)
        if min(curr) > max_d:
            return max_d + 1
        prev = curr
    return prev[lb]


# ---------------------------------------------------------------------------
# Main entry point
# ---------------------------------------------------------------------------

def audit_agent_config(
    *,
    repo: Optional[str] = None,
    project_path: Optional[str] = None,
    storage_path: Optional[str] = None,
) -> dict[str, Any]:
    """Audit agent configuration files for token waste and stale references.

    Args:
        repo: Repository identifier for cross-referencing symbols/files.
              If omitted, skips stale-reference and dead-path checks.
        project_path: Project directory to scan for config files.
                     Defaults to cwd.
        storage_path: Override index storage path.

    Returns:
        Structured audit result with files_scanned, total_tokens,
        per_turn_cost, findings, and token_breakdown.
    """
    project = project_path or os.getcwd()
    files = _discover_files(project)

    if not files:
        return {
            "files_scanned": 0,
            "total_tokens": 0,
            "per_turn_cost": "No agent config files found",
            "findings": [],
            "token_breakdown": [],
        }

    # Load index if repo is provided
    index = None
    all_symbol_names: set[str] = set()
    source_files: set[str] = set()
    source_root = ""

    if repo:
        try:
            from ._utils import resolve_repo as _resolve
            owner, name = _resolve(repo, storage_path)
            store = IndexStore(base_path=storage_path)
            index = store.load_index(owner, name)
            if index:
                all_symbol_names = {s.get("name", "") for s in index.symbols if s.get("name")}
                source_files = set(index.source_files)
                source_root = index.source_root or ""
        except Exception as e:
            logger.debug("Could not load index for %s: %s", repo, e)

    # Run all checks
    findings: list[dict[str, Any]] = []
    for f in files:
        findings.extend(_check_bloat(f))
        findings.extend(_check_scope_leaks(f))
        if index:
            findings.extend(_check_stale_symbols(f, index, all_symbol_names))
            findings.extend(_check_dead_paths(f, source_files, source_root))
    findings.extend(_check_redundancy(files))

    # Sort: warnings first, then info
    severity_order = {"error": 0, "warning": 1, "info": 2}
    findings.sort(key=lambda f: severity_order.get(f["severity"], 9))

    total_tokens = sum(f["tokens"] for f in files)
    global_tokens = sum(f["tokens"] for f in files if f["scope"] == "global")

    return {
        "files_scanned": len(files),
        "total_tokens": total_tokens,
        "global_tokens": global_tokens,
        "per_turn_cost": (
            f"~{total_tokens:,} tokens injected into every agent turn"
            + (f" ({global_tokens:,} from global config, applied to all projects)" if global_tokens else "")
        ),
        "findings": findings,
        "finding_counts": {
            "warning": sum(1 for f in findings if f["severity"] == "warning"),
            "info": sum(1 for f in findings if f["severity"] == "info"),
        },
        "token_breakdown": [
            {
                "file": f["path"],
                "scope": f["scope"],
                "tokens": f["tokens"],
                "lines": f["lines"],
                "description": f["description"],
            }
            for f in sorted(files, key=lambda x: x["tokens"], reverse=True)
        ],
    }
