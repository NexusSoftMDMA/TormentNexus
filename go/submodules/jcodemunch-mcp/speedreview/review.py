#!/usr/bin/env python
"""
speedreview — AI code review in seconds.

Runs inside a GitHub Action. Uses jCodeMunch for token-efficient code retrieval
and Groq for ultra-fast inference. Posts a structured review as a PR comment.

Environment variables (set by action.yml):
  GROQ_API_KEY           — Groq API key
  SPEEDREVIEW_MODEL      — Groq model (default: llama-3.3-70b-versatile)
  SPEEDREVIEW_SEVERITY   — Minimum severity: low, medium, high
  SPEEDREVIEW_MAX_LENGTH — Max PR comment length in characters
  SPEEDREVIEW_TOKEN_BUDGET — Token budget for jCodeMunch retrieval
  SPEEDREVIEW_BASE_REF   — Base ref to diff against (auto-detect if empty)
  GITHUB_TOKEN           — GitHub token for posting comments
  PR_NUMBER              — Pull request number
  GITHUB_REPOSITORY      — owner/repo
  GITHUB_WORKSPACE       — Checkout path
"""

from __future__ import annotations

import json
import os
import subprocess
import sys
import time

# ── Config ───────────────────────────────────────────────────────────────────

GROQ_API_KEY = os.environ.get("GROQ_API_KEY", "")
MODEL = os.environ.get("SPEEDREVIEW_MODEL", "llama-3.3-70b-versatile")
SEVERITY = os.environ.get("SPEEDREVIEW_SEVERITY", "low")
MAX_LENGTH = int(os.environ.get("SPEEDREVIEW_MAX_LENGTH", "4000"))
TOKEN_BUDGET = int(os.environ.get("SPEEDREVIEW_TOKEN_BUDGET", "8000"))
BASE_REF = os.environ.get("SPEEDREVIEW_BASE_REF", "")
GITHUB_TOKEN = os.environ.get("GITHUB_TOKEN", "")
PR_NUMBER = os.environ.get("PR_NUMBER", "")
REPO = os.environ.get("GITHUB_REPOSITORY", "")
WORKSPACE = os.environ.get("GITHUB_WORKSPACE", os.getcwd())

SEVERITY_ORDER = {"low": 0, "medium": 1, "high": 2}


# ── Git helpers ──────────────────────────────────────────────────────────────

def run_git(*args: str) -> str:
    result = subprocess.run(
        ["git", *args],
        capture_output=True, text=True, cwd=WORKSPACE,
        stdin=subprocess.DEVNULL,
    )
    if result.returncode != 0:
        print(f"::warning::git {' '.join(args)} failed: {result.stderr.strip()}")
        return ""
    return result.stdout.strip()


def get_base_ref() -> str:
    """Determine the base ref to diff against."""
    if BASE_REF:
        return BASE_REF
    # Try PR base from GitHub event
    event_path = os.environ.get("GITHUB_EVENT_PATH", "")
    if event_path and os.path.exists(event_path):
        with open(event_path) as f:
            event = json.load(f)
        base = event.get("pull_request", {}).get("base", {}).get("sha", "")
        if base:
            return base
    # Fallback: merge base with main/master
    for branch in ("origin/main", "origin/master"):
        merge_base = run_git("merge-base", branch, "HEAD")
        if merge_base:
            return merge_base
    return "HEAD~1"


def get_diff(base: str) -> str:
    """Get the unified diff for the PR."""
    return run_git("diff", base, "HEAD", "--", ".", ":(exclude)*.lock", ":(exclude)*.min.*")


def get_changed_files(base: str) -> list[str]:
    """Get list of changed file paths."""
    output = run_git("diff", "--name-only", "--diff-filter=ACMR", base, "HEAD")
    return [f for f in output.split("\n") if f.strip()] if output else []


# ── jCodeMunch retrieval ─────────────────────────────────────────────────────

def index_and_retrieve(changed_files: list[str], diff_text: str) -> dict:
    """Index the repo and retrieve relevant context using jCodeMunch."""
    from jcodemunch_mcp.tools.index_folder import index_folder
    from jcodemunch_mcp.tools.get_changed_symbols import get_changed_symbols
    from jcodemunch_mcp.tools.get_ranked_context import get_ranked_context

    context = {
        "changed_symbols": [],
        "blast_radius": [],
        "ranked_context": "",
    }

    # Index the workspace
    print("::group::Indexing repository")
    t0 = time.perf_counter()
    index_result = index_folder(path=WORKSPACE, use_ai_summaries=False)
    index_time = time.perf_counter() - t0
    if "error" in index_result:
        print(f"::warning::Index failed: {index_result['error']}")
        print("::endgroup::")
        return context

    total_symbols = index_result.get("total_symbols", "?")
    print(f"Indexed {total_symbols} symbols in {index_time:.1f}s")
    print("::endgroup::")

    # Resolve the repo identifier from the index result
    repo_id = index_result.get("repo", index_result.get("repo_id", ""))
    if not repo_id:
        # Fallback: use the workspace directory name
        repo_id = os.path.basename(WORKSPACE)

    # Get changed symbols with blast radius
    print("::group::Analyzing changes")
    t0 = time.perf_counter()
    base = get_base_ref()
    changes = get_changed_symbols(
        repo=repo_id,
        since_sha=base,
        include_blast_radius=True,
        max_blast_depth=2,
    )
    change_time = time.perf_counter() - t0

    if "error" not in changes:
        for category in ("changed_symbols", "added_symbols", "removed_symbols"):
            symbols = changes.get(category, [])
            if isinstance(symbols, list):
                context["changed_symbols"].extend(symbols)

        blast = changes.get("blast_radius", {})
        if isinstance(blast, dict):
            for sym_name, radius_info in blast.items():
                confirmed = radius_info.get("confirmed", []) if isinstance(radius_info, dict) else []
                if confirmed:
                    context["blast_radius"].append({
                        "symbol": sym_name,
                        "affected_files": len(confirmed),
                        "files": confirmed[:5],
                    })

        sym_count = len(context["changed_symbols"])
        blast_count = len(context["blast_radius"])
        print(f"Found {sym_count} changed symbols, {blast_count} with downstream impact ({change_time:.1f}s)")
    else:
        print(f"::warning::get_changed_symbols: {changes.get('error', 'unknown error')}")

    # Get ranked context for the changes
    t0 = time.perf_counter()
    # Build a query from changed symbol names + changed file names
    sym_names = [s.get("name", s) if isinstance(s, dict) else str(s)
                 for s in context["changed_symbols"][:10]]
    query_parts = sym_names + [os.path.basename(f) for f in changed_files[:5]]
    query = " ".join(query_parts[:15]) if query_parts else "main changes"

    ranked = get_ranked_context(
        repo=repo_id,
        query=query,
        token_budget=TOKEN_BUDGET,
        strategy="combined",
    )
    ranked_time = time.perf_counter() - t0

    if "error" not in ranked:
        items = ranked.get("context_items", [])
        # Format context items into readable text
        context_parts = []
        for item in items:
            header = f"# {item.get('file', '?')}::{item.get('name', '?')} ({item.get('kind', '?')})"
            source = item.get("source", "")
            if source:
                context_parts.append(f"{header}\n{source}")
        context["ranked_context"] = "\n\n".join(context_parts)
        print(f"Retrieved {len(items)} context items ({ranked_time:.1f}s)")
    else:
        print(f"::warning::get_ranked_context: {ranked.get('error', 'unknown error')}")

    print("::endgroup::")
    return context


# ── Groq inference ───────────────────────────────────────────────────────────

SYSTEM_PROMPT = """\
You are speedreview, an expert code reviewer. You receive a git diff and \
relevant code context from the repository. Produce a concise, actionable review.

Rules:
- Focus on bugs, security issues, performance problems, and logic errors.
- Do NOT comment on style, formatting, naming conventions, or missing docs.
- Each issue must reference a specific file and line number from the diff.
- Classify each issue as [High], [Medium], or [Low] severity.
- If no real issues are found, say so briefly — don't invent problems.
- Keep the summary to 1-2 sentences.
- Keep the total review under {max_length} characters.

Output format (markdown):
## speedreview ({{time}})

### Summary
1-2 sentence summary of what the PR does.

### Issues Found
- **[Severity]** Description (file:line)

### Impact Analysis
Brief note on downstream effects based on the blast radius data.

---
*Powered by [jCodeMunch](https://github.com/jgravelle/jcodemunch-mcp) + \
[Groq](https://groq.com) | Review completed in {{time}}*
"""


def call_groq(diff_text: str, context: dict, elapsed_so_far: float) -> str:
    """Send the assembled context + diff to Groq and return the review."""
    from openai import OpenAI

    client = OpenAI(api_key=GROQ_API_KEY, base_url="https://api.groq.com/openai/v1")

    # Build the user message
    parts = []

    # Changed symbols summary
    if context["changed_symbols"]:
        sym_lines = []
        for s in context["changed_symbols"][:20]:
            if isinstance(s, dict):
                sym_lines.append(f"- {s.get('change', '?')}: {s.get('name', '?')} in {s.get('file', '?')}")
            else:
                sym_lines.append(f"- {s}")
        parts.append("## Changed Symbols\n" + "\n".join(sym_lines))

    # Blast radius
    if context["blast_radius"]:
        blast_lines = []
        for b in context["blast_radius"][:10]:
            files = ", ".join(os.path.basename(f) if isinstance(f, str) else str(f)
                             for f in b.get("files", [])[:3])
            blast_lines.append(f"- `{b['symbol']}` affects {b['affected_files']} files: {files}")
        parts.append("## Blast Radius (downstream impact)\n" + "\n".join(blast_lines))

    # Ranked context (surrounding code)
    if context["ranked_context"]:
        # Truncate if too long
        rc = context["ranked_context"]
        if len(rc) > 12000:
            rc = rc[:12000] + "\n\n... (truncated)"
        parts.append("## Relevant Code Context\n" + rc)

    # The diff itself (truncate very large diffs)
    diff_section = diff_text
    if len(diff_section) > 15000:
        diff_section = diff_section[:15000] + "\n\n... (diff truncated)"
    parts.append("## Git Diff\n```diff\n" + diff_section + "\n```")

    user_message = "\n\n".join(parts)
    system = SYSTEM_PROMPT.format(max_length=MAX_LENGTH)

    t0 = time.perf_counter()
    response = client.chat.completions.create(
        model=MODEL,
        messages=[
            {"role": "system", "content": system},
            {"role": "user", "content": user_message},
        ],
        temperature=0.2,
        max_tokens=2000,
    )
    inference_time = time.perf_counter() - t0

    review = response.choices[0].message.content or ""

    # Replace {{time}} placeholder with actual total time
    total_time = elapsed_so_far + inference_time
    time_str = f"{total_time:.1f}s"
    review = review.replace("{{time}}", time_str)

    # Also fix the header if the model didn't use the placeholder
    if "speedreview (" not in review:
        review = f"## speedreview ({time_str})\n\n{review}"

    return review, total_time


# ── Severity filtering ───────────────────────────────────────────────────────

def filter_by_severity(review: str) -> str:
    """Remove issues below the severity threshold."""
    if SEVERITY == "low":
        return review
    threshold = SEVERITY_ORDER.get(SEVERITY, 0)
    lines = review.split("\n")
    filtered = []
    for line in lines:
        # Check if this is an issue line with a severity tag
        is_issue = line.strip().startswith("- **[")
        if is_issue:
            line_lower = line.lower()
            if "[high]" in line_lower and threshold <= 2:
                filtered.append(line)
            elif "[medium]" in line_lower and threshold <= 1:
                filtered.append(line)
            elif "[low]" in line_lower and threshold <= 0:
                filtered.append(line)
            # Skip issues below threshold
        else:
            filtered.append(line)
    return "\n".join(filtered)


# ── GitHub comment ───────────────────────────────────────────────────────────

def post_comment(body: str):
    """Post or update a PR comment via gh CLI."""
    if not PR_NUMBER or not GITHUB_TOKEN:
        print("::warning::No PR_NUMBER or GITHUB_TOKEN — printing review to stdout")
        print(body)
        return

    # Check for existing speedreview comment to update
    existing = subprocess.run(
        ["gh", "api", f"repos/{REPO}/issues/{PR_NUMBER}/comments",
         "--jq", '.[] | select(.body | startswith("## speedreview")) | .id'],
        capture_output=True, text=True,
        stdin=subprocess.DEVNULL,
        env={**os.environ, "GH_TOKEN": GITHUB_TOKEN},
    )

    comment_ids = [cid.strip() for cid in existing.stdout.strip().split("\n") if cid.strip()]

    if comment_ids:
        # Update existing comment
        comment_id = comment_ids[0]
        subprocess.run(
            ["gh", "api", "--method", "PATCH",
             f"repos/{REPO}/issues/comments/{comment_id}",
             "-f", f"body={body}"],
            capture_output=True, text=True,
            stdin=subprocess.DEVNULL,
            env={**os.environ, "GH_TOKEN": GITHUB_TOKEN},
        )
        print(f"Updated existing speedreview comment #{comment_id}")
    else:
        # Create new comment
        subprocess.run(
            ["gh", "pr", "comment", PR_NUMBER, "--body", body],
            capture_output=True, text=True,
            stdin=subprocess.DEVNULL,
            env={**os.environ, "GH_TOKEN": GITHUB_TOKEN},
        )
        print("Posted new speedreview comment")


# ── Main ─────────────────────────────────────────────────────────────────────

def main():
    if not GROQ_API_KEY:
        sys.exit("GROQ_API_KEY environment variable is required")

    wall_start = time.perf_counter()

    print("speedreview — AI code review powered by jCodeMunch + Groq")
    print(f"  Model: {MODEL}")
    print(f"  Severity threshold: {SEVERITY}")
    print(f"  Token budget: {TOKEN_BUDGET}")
    print()

    # Get the diff
    base = get_base_ref()
    print(f"Base ref: {base}")

    changed_files = get_changed_files(base)
    if not changed_files:
        print("No changed files detected — skipping review.")
        return

    print(f"Changed files: {len(changed_files)}")
    for f in changed_files[:10]:
        print(f"  {f}")
    if len(changed_files) > 10:
        print(f"  ... and {len(changed_files) - 10} more")
    print()

    diff_text = get_diff(base)
    if not diff_text:
        print("::warning::Empty diff — skipping review.")
        return

    # Retrieve context via jCodeMunch
    context = index_and_retrieve(changed_files, diff_text)

    retrieval_time = time.perf_counter() - wall_start

    # Call Groq for the review
    print("::group::Generating review via Groq")
    review, total_time = call_groq(diff_text, context, retrieval_time)
    print(f"Total time: {total_time:.1f}s (retrieval: {retrieval_time:.1f}s, inference: {total_time - retrieval_time:.1f}s)")
    print("::endgroup::")

    # Filter by severity
    review = filter_by_severity(review)

    # Truncate if needed
    if len(review) > MAX_LENGTH:
        review = review[:MAX_LENGTH - 100] + "\n\n... (truncated)\n\n---\n*Review truncated at character limit*"

    # Post the comment
    post_comment(review)

    print(f"\nspeedreview complete in {total_time:.1f}s")


if __name__ == "__main__":
    main()
