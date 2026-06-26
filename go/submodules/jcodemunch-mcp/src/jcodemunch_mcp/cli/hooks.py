"""Claude Code hook handlers for jCodemunch enforcement.

PreToolUse  — two nudges toward jCodemunch before a native file tool runs:
              Read on a large code file → get_file_outline / get_symbol_source;
              Grep inside an indexed repo → search_text / search_symbols /
              find_references (exhaust the credited jcm routes before a raw scan).
PostToolUse — auto-reindex after Edit/Write to keep the index fresh.

Both read JSON from stdin and write JSON to stdout per the Claude Code
hooks specification.
"""

import json
import os
import subprocess
import sys


# Extensions that benefit from jCodemunch structural navigation.
# Kept intentionally broad — mirrors languages.py LANGUAGE_REGISTRY.
_CODE_EXTENSIONS: set[str] = {
    ".py", ".pyi",
    ".js", ".jsx", ".mjs", ".cjs",
    ".ts", ".tsx", ".mts", ".cts",
    ".go",
    ".rs",
    ".java",
    ".php",
    ".rb",
    ".cs", ".cshtml", ".razor",
    ".cpp", ".c", ".h", ".hpp", ".cc", ".cxx", ".ino", ".pde",
    ".vhd", ".vhdl", ".vho", ".vhs",
    ".v", ".vh", ".sv", ".svh",
    ".swift",
    ".kt", ".kts",
    ".scala",
    ".dart",
    ".lua", ".luau",
    ".ex", ".exs",
    ".erl", ".hrl",
    ".vue", ".astro", ".svelte",
    ".sql",
    ".gd",       # GDScript
    ".al",       # AL (Business Central)
    ".gleam",
    ".nix",
    ".hcl", ".tf",
    ".proto",
    ".graphql", ".gql",
    ".verse",
    ".jl",       # Julia
    ".r", ".R",
    ".hs",       # Haskell
    ".f90", ".f95", ".f03", ".f08",  # Fortran
    ".groovy",
    ".pl", ".pm",  # Perl
    ".bash", ".sh", ".zsh",
}

# Minimum file size to trigger jCodemunch suggestion.
# Override with JCODEMUNCH_HOOK_MIN_SIZE env var.
_MIN_SIZE_BYTES = int(os.environ.get("JCODEMUNCH_HOOK_MIN_SIZE", "4096"))


def _enforce_mode() -> str:
    """jCodemunch enforcement tier for native file tools (``JCODEMUNCH_ENFORCE``).

    * ``"advisory"`` (default): warn on stderr but **allow** — the v1.108.47
      behavior. A hard deny here would break Read-before-Edit.
    * ``"strict"``: **deny** a native Read/Grep that an indexed-repo jcm route
      can already serve. Targeted reads (``offset``/``limit``), tiny files, and
      paths outside every indexed repo still pass, so the escape hatch is always
      one step away and jcm is never blamed for a search it can't serve.
    * ``"off"``: no nudge, no deny — fully silent.

    Unknown values fall back to ``"advisory"`` so a typo never hard-blocks tools.
    Opt in to strict with ``jcodemunch-mcp init --strict`` (persists the env var
    into ~/.claude/settings.json) or by exporting it yourself.
    """
    val = os.environ.get("JCODEMUNCH_ENFORCE", "advisory").strip().lower()
    if val in {"strict", "deny", "block", "hard"}:
        return "strict"
    if val in {"off", "0", "false", "no", "none", "silent"}:
        return "off"
    return "advisory"


def _emit_pretooluse_deny(reason: str) -> int:
    """Emit a Claude Code PreToolUse ``deny`` decision (stdout JSON) and exit 0.

    The deny lives in the JSON decision channel; stderr stays free for advisory
    hints, and exit code stays 0 so the harness reads the decision, not a crash.
    """
    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "permissionDecision": "deny",
            "permissionDecisionReason": reason,
        }
    }))
    return 0


def run_pretooluse() -> int:
    """PreToolUse hook: steer Claude toward jCodemunch before native file tools.

    Dispatches on the tool being called:
      * ``Grep`` whose search root falls inside an indexed repo → the credited
        jcm routes (``search_text`` / ``search_symbols`` / ``find_references``)
        serve the same scan, ranked and on the savings meter.
      * ``Read`` on a code file above the size threshold → ``get_file_outline``
        + ``get_symbol_source`` instead.

    Strength is set by ``_enforce_mode()`` (``JCODEMUNCH_ENFORCE``): the default
    ``advisory`` tier warns on stderr but allows (so Read-before-Edit and the
    Grep fallback keep working); ``strict`` denies the same calls but only when
    an indexed-repo jcm route can serve them — targeted reads (``offset`` /
    ``limit``), tiny files, non-code files, and paths outside every indexed repo
    always pass; ``off`` is fully silent.

    Returns exit code (always 0 — errors are swallowed to avoid blocking).
    """
    try:
        data = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0  # Unparseable → allow

    mode = _enforce_mode()
    if mode == "off":
        return 0  # No nudge, no deny.

    tool_input = data.get("tool_input", {}) or {}

    # Grep is the tool that most often does a job a jcm route already covers,
    # entirely off the savings meter — intercept it first. Defensive: the hook
    # must never crash the agent's Grep, so swallow anything unexpected.
    if data.get("tool_name", "") == "Grep":
        try:
            if mode == "strict":
                return _strict_grep(tool_input, data.get("cwd", ""))
            return _nudge_grep(tool_input, data.get("cwd", ""))
        except Exception:
            return 0

    # --- Read interception ---
    file_path: str = tool_input.get("file_path", "")
    if not file_path:
        return 0

    # Check extension
    _, ext = os.path.splitext(file_path)
    if ext.lower() not in _CODE_EXTENSIONS:
        return 0  # Not a code file → allow

    # Check size
    try:
        size = os.path.getsize(file_path)
    except OSError:
        return 0  # Can't stat → allow (file may not exist yet)

    if size < _MIN_SIZE_BYTES:
        return 0  # Small file → allow

    # Targeted reads (offset/limit set) are likely pre-edit — allow silently.
    if tool_input.get("offset") is not None or tool_input.get("limit") is not None:
        return 0

    if mode == "strict":
        # Strict: deny the full-file read, but only when the file lives inside an
        # indexed repo (jcm can actually serve it). Reads outside every indexed
        # repo pass — denying them would block work jcm can't help with.
        try:
            roots = _indexed_source_roots()
            norm = os.path.normcase(os.path.abspath(file_path))
            if roots and _path_overlaps(norm, roots):
                return _emit_pretooluse_deny(
                    f"jCodemunch strict mode: this {size:,}-byte code file is in an "
                    "indexed repo. Use get_file_outline + get_symbol_source instead "
                    "of a full Read. For an exact-line pre-Edit read, pass "
                    "offset/limit and it will pass. (JCODEMUNCH_ENFORCE=advisory "
                    "for warn-only, =off to disable.)"
                )
        except Exception:
            return 0  # Never block on a store hiccup.
        return 0

    # Advisory: full-file exploratory read on a large code file — warn but allow.
    # Hard deny breaks the Edit workflow (Claude Code requires Read before Edit).
    # Stderr text is surfaced to the agent as guidance.
    print(
        f"jCodemunch hint: this is a {size:,}-byte code file. "
        "Prefer get_file_outline + get_symbol_source for exploration. "
        "Use Read only when you need exact line numbers for Edit.",
        file=sys.stderr,
    )
    return 0


# ---------------------------------------------------------------------------
# Grep → jCodemunch nudge (PreToolUse, matcher "Grep")
# ---------------------------------------------------------------------------

def _grep_nudge_enabled() -> bool:
    """Whether the Grep→jcm nudge is active. Set JCODEMUNCH_HOOK_GREP_NUDGE=0
    (or false/no/off) to silence it."""
    return os.environ.get("JCODEMUNCH_HOOK_GREP_NUDGE", "1").strip().lower() not in {
        "0", "false", "no", "off",
    }


def _indexed_source_roots() -> list[str]:
    """Normalised absolute source roots of every locally-indexed repo.

    Loaded fresh per call (the hook is a short-lived process). Best-effort:
    any failure yields ``[]`` so the hook silently allows the Grep rather than
    ever blocking it on a store hiccup.
    """
    try:
        from ..storage import IndexStore
        roots: list[str] = []
        for entry in IndexStore().list_repos():
            sr = (entry.get("source_root") or "").strip()
            if sr:
                roots.append(os.path.normcase(os.path.abspath(sr)))
        return roots
    except Exception:
        return []


def _grep_search_root(tool_input: dict, cwd: str) -> str:
    """Resolve the directory a Grep call will actually scan (normalised)."""
    path = (tool_input.get("path") or "").strip()
    base = cwd or os.getcwd()
    if not path:
        root = base
    elif os.path.isabs(path):
        root = path
    else:
        root = os.path.join(base, path)
    return os.path.normcase(os.path.abspath(root))


def _path_overlaps(root: str, source_roots: list[str]) -> bool:
    """True when *root* is equal to, inside, or an ancestor of any indexed root.

    The ancestor case matters too: grepping a parent directory that *contains*
    an indexed repo is still a search jcm can serve.
    """
    for sr in source_roots:
        if root == sr or root.startswith(sr + os.sep) or sr.startswith(root + os.sep):
            return True
    return False


def _nudge_grep(tool_input: dict, cwd: str) -> int:
    """Grep PreToolUse branch: when the search targets an indexed repo, steer
    the agent to the credited jcm retrieval routes first. Allows the Grep
    regardless (exit 0) — this is a nudge, not a block."""
    if not _grep_nudge_enabled():
        return 0
    roots = _indexed_source_roots()
    if not roots:
        return 0  # Nothing indexed → jcm can't help here, stay silent.
    if not _path_overlaps(_grep_search_root(tool_input, cwd), roots):
        return 0  # Grep is outside every indexed repo → allow silently.

    pattern = (tool_input.get("pattern") or "").strip()
    for_pat = f" for `{pattern}`" if pattern else ""
    print(
        f"jCodemunch: this Grep{for_pat} targets an indexed repo. Exhaust the jcm "
        "routes first, since they're tighter and credited (raw Grep is neither):\n"
        "  - search_text     : same regex/substring scan, ranked + winnowed\n"
        "  - search_symbols  : when hunting a definition (function/class/const/type)\n"
        "  - find_references / find_importers : 'where is X used / who imports this'\n"
        "Fall back to Grep only once those come up empty.",
        file=sys.stderr,
    )
    return 0


def _strict_grep(tool_input: dict, cwd: str) -> int:
    """Strict-mode Grep branch: **deny** a Grep that targets an indexed repo so
    the agent uses the credited jcm routes. A Grep outside every indexed repo
    (jcm can't serve it) is allowed silently."""
    roots = _indexed_source_roots()
    if not roots or not _path_overlaps(_grep_search_root(tool_input, cwd), roots):
        return 0  # Nothing indexed / outside every repo → allow.
    pattern = (tool_input.get("pattern") or "").strip()
    for_pat = f" for `{pattern}`" if pattern else ""
    return _emit_pretooluse_deny(
        f"jCodemunch strict mode: this Grep{for_pat} targets an indexed repo. Use "
        "search_text (same scan, ranked + credited), search_symbols (definitions), "
        "or find_references / find_importers ('where is X used / who imports this') "
        "instead. (JCODEMUNCH_ENFORCE=advisory for warn-only, =off to disable.)"
    )


def run_posttooluse() -> int:
    """PostToolUse hook: auto-index files after Edit/Write.

    Reads hook JSON from stdin, extracts the file path, and spawns
    ``jcodemunch-mcp index-file <path>`` as a fire-and-forget background
    process to keep the index fresh.

    Non-code files are skipped.  Errors are swallowed silently.

    Returns exit code (always 0).
    """
    try:
        data = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    file_path: str = data.get("tool_input", {}).get("file_path", "")
    if not file_path:
        return 0

    # Only re-index code files
    _, ext = os.path.splitext(file_path)
    if ext.lower() not in _CODE_EXTENSIONS:
        return 0

    # Fire-and-forget: spawn index-file in background
    try:
        kwargs: dict = dict(
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        # On Windows, CREATE_NO_WINDOW prevents a console flash
        if sys.platform == "win32":
            kwargs["creationflags"] = subprocess.CREATE_NO_WINDOW  # type: ignore[attr-defined]
        subprocess.Popen(
            ["jcodemunch-mcp", "index-file", file_path],
            **kwargs,
        )
    except (OSError, FileNotFoundError):
        pass  # jcodemunch-mcp not in PATH → skip silently

    return 0


def run_copilot_posttooluse() -> int:
    """GitHub Copilot ``postToolUse`` hook: auto-index files after Edit/Write.

    Adapter for the Copilot CLI / cloud-agent hook payload shape, which
    differs from Claude Code's:

    Copilot stdin JSON::

        {
            "timestamp": "...",
            "cwd": "...",
            "toolName": "edit" | "write" | "create_file" | ...,
            "toolArgs": "{\\"path\\": \\"/abs/path/to/file.py\\", ...}",
            "toolResult": "..."
        }

    ``toolArgs`` arrives as a JSON-encoded **string**, not a nested object.
    Tool names vary across Copilot tool implementations, so we extract a
    file path heuristically: any value at the top level of toolArgs whose
    key matches ``path``/``file_path``/``filename``/``filePath`` and points
    at an existing file. If the file is a code file under a directory that
    has been indexed, spawn ``jcodemunch-mcp index-file <path>`` as a
    fire-and-forget background process. Errors are swallowed silently —
    Copilot ignores postToolUse stdout/exit code, so a failing reindex
    must never disrupt the agent flow.
    """
    try:
        data = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    tool_args_raw = data.get("toolArgs", "")
    if isinstance(tool_args_raw, str):
        try:
            tool_args = json.loads(tool_args_raw) if tool_args_raw else {}
        except (json.JSONDecodeError, ValueError):
            return 0
    elif isinstance(tool_args_raw, dict):
        tool_args = tool_args_raw
    else:
        return 0

    file_path = ""
    for key in ("file_path", "filePath", "path", "filename"):
        v = tool_args.get(key)
        if isinstance(v, str) and v:
            file_path = v
            break
    if not file_path:
        return 0

    _, ext = os.path.splitext(file_path)
    if ext.lower() not in _CODE_EXTENSIONS:
        return 0

    try:
        kwargs: dict = dict(
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        if sys.platform == "win32":
            kwargs["creationflags"] = subprocess.CREATE_NO_WINDOW  # type: ignore[attr-defined]
        subprocess.Popen(
            ["jcodemunch-mcp", "index-file", file_path],
            **kwargs,
        )
    except (OSError, FileNotFoundError):
        pass

    return 0


def run_precompact() -> int:
    """PreCompact hook: generate session snapshot before context compaction.

    Reads hook JSON from stdin. Builds a compact snapshot of the current
    session state and returns it as a message for context injection.

    Returns exit code (always 0 — errors are swallowed to avoid blocking).
    """
    try:
        json.load(sys.stdin)  # Validate stdin is valid JSON
    except (json.JSONDecodeError, ValueError):
        return 0

    # The hook runs as a SEPARATE process from the MCP server, so the in-process
    # SessionJournal is empty (#334). Read the live journal the server persists
    # incrementally first; fall back to the in-process journal (covers embedded
    # invocations); finally emit an explicit "no live session" message rather
    # than a misleading zero-state snapshot.
    snapshot_text = ""
    live_context = None
    try:
        from jcodemunch_mcp.tools.get_session_snapshot import snapshot_from_live
        live = snapshot_from_live()
        if live:
            snapshot_text = live.get("snapshot", "")
            live_context = live.get("_context")
    except Exception:
        snapshot_text = ""

    if not snapshot_text:
        try:
            from jcodemunch_mcp.tools.get_session_snapshot import get_session_snapshot
            snap = get_session_snapshot()
            structured = snap.get("structured", {})
            if structured.get("total_files_explored") or structured.get("total_searches"):
                snapshot_text = snap.get("snapshot", "")
        except Exception:
            snapshot_text = ""

    used_fallback = False
    if not snapshot_text:
        # Honest fallback — never emit a healthy-looking zero-state snapshot.
        snapshot_text = _precompact_no_journal_message()
        used_fallback = True

    # Enrich with structural landmarks (PageRank top-N) and recently-changed
    # symbols. Seed from the live journal context when we have one so landmarks
    # work out-of-process too; skip entirely on the no-journal fallback.
    if not used_fallback:
        try:
            landmarks = _build_landmark_section(context=live_context)
            if landmarks:
                snapshot_text += landmarks
        except Exception:
            pass  # Landmark enrichment must not block compaction

    # Return snapshot as hook output for context injection.
    # PreCompact has no hookSpecificOutput variant in Claude Code's schema,
    # so we use the top-level systemMessage field instead.
    result = {
        "systemMessage": snapshot_text,
    }
    json.dump(result, sys.stdout)
    return 0


def _precompact_no_journal_message() -> str:
    """Explicit fallback when no live jCodeMunch session journal is readable.

    The PreCompact hook runs in a separate process from the MCP server (#334);
    when the server has not persisted a live journal this session, say so plainly
    instead of emitting a `Files explored: 0` snapshot that looks successful.
    """
    return (
        "## Session Snapshot (jCodemunch)\n"
        "No live jCodeMunch session journal was available to this PreCompact hook "
        "process. The hook runs separately from the MCP server, so it could not "
        "read in-flight session state. If jCodeMunch tools were used this session, "
        "their snapshot is unavailable here; otherwise no code exploration was "
        "recorded. (Set JCODEMUNCH_LIVE_JOURNAL=0 only if you intend to disable "
        "the live journal.)"
    )


# ---------------------------------------------------------------------------
# Landmark enrichment helpers (Gap 4A — Structural Landmarks)
# ---------------------------------------------------------------------------

def _build_landmark_section(top_n: int = 20, context: "dict | None" = None) -> str:
    """Build a compact landmarks + recently-changed section for PreCompact.

    Queries all indexed repos visible in the session journal's edited files,
    computes PageRank to find the most structurally central symbols, and
    cross-references the journal's edit log to surface recently-changed symbols.

    When ``context`` is supplied (e.g. the live journal read by the
    out-of-process hook, #334) it is used instead of the empty in-process
    journal. Returns a markdown string to append to the snapshot, or "" if no
    data.
    """
    import logging
    logger = logging.getLogger(__name__)

    try:
        from ..storage import IndexStore
        from ..tools.pagerank import compute_pagerank
        from ..tools.session_journal import get_journal
    except Exception:
        logger.debug("landmark imports failed", exc_info=True)
        return ""

    if context is None:
        journal = get_journal()
        context = journal.get_context(max_files=50, max_queries=0, max_edits=50)
    edited_files = [e["file"] for e in context.get("files_edited", [])]
    accessed_files = [f["file"] for f in context.get("files_accessed", [])]

    if not edited_files and not accessed_files:
        return ""

    # Load all indexed repos and find which ones contain session files
    store = IndexStore()
    repo_indices: dict[str, object] = {}
    try:
        repos = store.list_repos()
    except Exception:
        return ""

    for entry in repos:
        owner = entry.get("owner", "")
        name = entry.get("name", "")
        if not owner or not name:
            continue
        repo_id = f"{owner}/{name}"
        if repo_id in repo_indices:
            continue
        try:
            idx = store.load_index(owner, name)
            if idx and idx.source_files:
                repo_indices[repo_id] = idx
        except Exception:
            continue

    if not repo_indices:
        return ""

    parts: list[str] = []

    for repo_id, index in repo_indices.items():
        if not index.imports or not index.source_files:
            continue

        # Compute PageRank
        try:
            pr_scores, _ = compute_pagerank(
                index.imports, index.source_files,
                alias_map=getattr(index, "alias_map", None),
                psr4_map=getattr(index, "psr4_map", None),
            )
        except Exception:
            logger.debug("PageRank failed for %s", repo_id, exc_info=True)
            continue

        if not pr_scores:
            continue

        # Rank files by PageRank, then pick top symbols from those files
        top_files = sorted(pr_scores.items(), key=lambda x: x[1], reverse=True)[:top_n * 2]
        top_file_set = {f for f, _ in top_files}

        # Collect symbols from top-ranked files
        symbol_pr: list[tuple[dict, float]] = []
        for sym in index.symbols:
            f = sym.get("file", "")
            if f in top_file_set:
                symbol_pr.append((sym, pr_scores.get(f, 0.0)))

        # Sort by PageRank score, take top_n
        symbol_pr.sort(key=lambda x: x[1], reverse=True)
        landmarks = symbol_pr[:top_n]

        if landmarks:
            parts.append(f"\n\n### Structural Landmarks ({repo_id})")
            for sym, score in landmarks:
                name = sym.get("name", "?")
                kind = sym.get("kind", "")
                f = sym.get("file", "")
                line = sym.get("line", 0)
                summary = sym.get("summary", "")
                loc = f"{f}:{line}" if line else f
                desc = f" — {summary}" if summary else ""
                parts.append(f"- `{name}` ({kind}, {loc}){desc}")

        # Recently-changed symbols: cross-ref edited files with index
        session_edited = {ef for ef in edited_files}
        changed_syms: list[dict] = []
        for sym in index.symbols:
            if sym.get("file", "") in session_edited:
                changed_syms.append(sym)

        if changed_syms:
            parts.append(f"\n### Recently Changed ({repo_id})")
            # Deduplicate and limit
            seen: set[str] = set()
            count = 0
            for sym in changed_syms:
                sid = sym.get("id", sym.get("name", ""))
                if sid in seen:
                    continue
                seen.add(sid)
                name = sym.get("name", "?")
                kind = sym.get("kind", "")
                f = sym.get("file", "")
                line = sym.get("line", 0)
                parts.append(f"- `{name}` ({kind}, {f}:{line})")
                count += 1
                if count >= 20:
                    break

    return "\n".join(parts) if parts else ""


# ---------------------------------------------------------------------------
# Post-task diagnostics hook (Gap 4B)
# ---------------------------------------------------------------------------

def run_taskcomplete() -> int:
    """TaskCompleted hook: surface dead code, untested symbols, and dangling refs.

    Reads hook JSON from stdin. Inspects files modified during the session
    and runs three diagnostic checks scoped to those files:
      1. find_dead_code — newly-orphaned symbols
      2. get_untested_symbols — new code with no test reachability
      3. check_references — dangling references to deleted/renamed symbols

    Returns exit code (always 0 — errors are swallowed to avoid blocking).
    """
    try:
        json.load(sys.stdin)  # Validate stdin
    except (json.JSONDecodeError, ValueError):
        return 0

    try:
        from ..tools.session_journal import get_journal
        journal = get_journal()
        context = journal.get_context(max_files=50, max_queries=0, max_edits=50)
    except Exception:
        return 0

    edited_files = [e["file"] for e in context.get("files_edited", [])]
    if not edited_files:
        return 0  # Nothing modified — nothing to diagnose

    # Find which repos contain these files
    try:
        from ..storage import IndexStore
        store = IndexStore()
        repos = store.list_repos()
    except Exception:
        return 0

    diagnostics: list[dict] = []

    for entry in repos:
        owner = entry.get("owner", "")
        name = entry.get("name", "")
        if not owner or not name:
            continue
        repo_id = f"{owner}/{name}"

        try:
            idx = store.load_index(owner, name)
        except Exception:
            continue
        if not idx or not idx.source_files:
            continue

        # Scope: only files in this repo that were edited
        repo_files = set(idx.source_files)
        session_files = [f for f in edited_files if f in repo_files]
        if not session_files:
            continue

        diag: dict = {"repo": repo_id, "files_checked": len(session_files)}

        # 1. Dead code scoped to edited files
        try:
            from ..tools.find_dead_code import find_dead_code
            dead_result = find_dead_code(repo_id, granularity="symbol")
            if dead_result and not dead_result.get("error"):
                dead_in_session = [
                    s for s in dead_result.get("dead_symbols", [])
                    if s.get("file") in set(session_files)
                ]
                if dead_in_session:
                    diag["dead_symbols"] = dead_in_session[:10]
        except Exception:
            pass

        # 2. Untested symbols in edited files
        try:
            from ..tools.get_untested_symbols import get_untested_symbols
            for sf in session_files[:5]:  # Limit to avoid slow scans
                # Convert file path to a glob pattern
                pattern = sf.replace("\\", "/")
                untested = get_untested_symbols(repo_id, file_pattern=pattern, max_results=5)
                if untested and not untested.get("error"):
                    syms = untested.get("untested_symbols", [])
                    if syms:
                        diag.setdefault("untested_symbols", []).extend(syms[:5])
        except Exception:
            pass

        # 3. Dangling references — check symbols that were in edited files
        try:
            from ..tools.check_references import check_references
            edited_syms = [
                sym["name"] for sym in idx.symbols
                if sym.get("file") in set(session_files)
            ][:10]
            if edited_syms:
                for sym_name in edited_syms:
                    ref_result = check_references(repo_id, identifier=sym_name, max_content_results=3)
                    if ref_result and not ref_result.get("error"):
                        if ref_result.get("total_references", 0) == 0:
                            diag.setdefault("unreferenced_symbols", []).append(sym_name)
        except Exception:
            pass

        if len(diag) > 2:  # More than just repo + files_checked
            diagnostics.append(diag)

    if not diagnostics:
        return 0

    # Build compact message for the agent
    parts = ["## Post-Task Diagnostics (jCodemunch)"]
    for diag in diagnostics:
        parts.append(f"\n### {diag['repo']} ({diag['files_checked']} files checked)")
        if "dead_symbols" in diag:
            parts.append(f"**Possibly orphaned:** {len(diag['dead_symbols'])} symbol(s)")
            for s in diag["dead_symbols"][:5]:
                parts.append(f"  - `{s.get('name', '?')}` ({s.get('file', '?')}:{s.get('line', 0)})")
        if "untested_symbols" in diag:
            parts.append(f"**No test coverage:** {len(diag['untested_symbols'])} symbol(s)")
            for s in diag["untested_symbols"][:5]:
                parts.append(f"  - `{s.get('name', '?')}` ({s.get('file', '?')})")
        if "unreferenced_symbols" in diag:
            parts.append(f"**Unreferenced:** {', '.join(f'`{s}`' for s in diag['unreferenced_symbols'][:5])}")

    result = {"systemMessage": "\n".join(parts)}
    json.dump(result, sys.stdout)
    return 0


# ---------------------------------------------------------------------------
# Subagent briefing hook (Gap 4C)
# ---------------------------------------------------------------------------

def run_subagentstart() -> int:
    """SubagentStart hook: inject condensed repo orientation for spawned agents.

    Reads hook JSON from stdin. Returns a compact briefing containing:
      - Repo stats (files, symbols, languages)
      - Top 15 structurally central symbols (PageRank)
      - Available jCodemunch tool catalog

    Returns exit code (always 0).
    """
    try:
        json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    try:
        from ..storage import IndexStore
        store = IndexStore()
        repos = store.list_repos()
    except Exception:
        return 0

    if not repos:
        return 0

    parts = ["## jCodemunch Repo Briefing"]

    for entry in repos:
        owner = entry.get("owner", "")
        name = entry.get("name", "")
        if not owner or not name:
            continue
        repo_id = f"{owner}/{name}"

        try:
            idx = store.load_index(owner, name)
        except Exception:
            continue
        if not idx:
            continue

        # Stats
        n_files = len(idx.source_files)
        n_symbols = len(idx.symbols)
        langs = set()
        for sym in idx.symbols:
            lang = sym.get("language")
            if lang:
                langs.add(lang)
        lang_str = ", ".join(sorted(langs)[:8]) if langs else "unknown"

        parts.append(f"\n### {repo_id}")
        parts.append(f"- **Files:** {n_files} | **Symbols:** {n_symbols} | **Languages:** {lang_str}")

        # Top central symbols via PageRank
        if idx.imports and idx.source_files:
            try:
                from ..tools.pagerank import compute_pagerank
                pr_scores, _ = compute_pagerank(
                    idx.imports, idx.source_files,
                    alias_map=getattr(idx, "alias_map", None),
                    psr4_map=getattr(idx, "psr4_map", None),
                )
                if pr_scores:
                    top_files = sorted(pr_scores.items(), key=lambda x: x[1], reverse=True)[:30]
                    top_file_set = {f for f, _ in top_files}
                    sym_pr = sorted(
                        [(sym, pr_scores.get(sym.get("file", ""), 0.0)) for sym in idx.symbols if sym.get("file", "") in top_file_set],
                        key=lambda x: x[1],
                        reverse=True,
                    )[:15]
                    if sym_pr:
                        parts.append("- **Key symbols:**")
                        for sym, _ in sym_pr:
                            parts.append(f"  - `{sym.get('name', '?')}` ({sym.get('kind', '')}, {sym.get('file', '')}:{sym.get('line', 0)})")
            except Exception:
                pass

    # Tool catalog (compact)
    parts.append("\n### Available jCodemunch Tools")
    parts.append(
        "search_symbols, get_symbol_source, get_context_bundle, get_file_content, "
        "search_text, get_ranked_context, find_importers, find_references, "
        "check_references, get_dependency_graph, get_class_hierarchy, "
        "get_call_hierarchy, get_blast_radius, get_impact_preview, "
        "get_changed_symbols, find_dead_code, get_untested_symbols, "
        "get_symbol_complexity, get_churn_rate, get_hotspots, get_repo_health, "
        "get_coupling_metrics, get_extraction_candidates, check_rename_safe, "
        "plan_refactoring, "
        "get_file_outline, get_file_tree, get_repo_outline, index_folder, "
        "index_repo, embed_repo, plan_turn, suggest_queries, "
        "get_session_context, get_session_snapshot, get_session_stats, "
        "get_cross_repo_map, get_layer_violations, audit_agent_config, "
        "get_dead_code_v2, search_columns"
    )
    parts.append("\nUse `plan_turn` to get recommended approach for your task.")

    result = {"systemMessage": "\n".join(parts)}
    json.dump(result, sys.stdout)
    return 0
