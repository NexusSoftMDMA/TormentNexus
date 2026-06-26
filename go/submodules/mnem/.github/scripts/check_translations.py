#!/usr/bin/env python3
"""
check_translations.py - verify translations stay in sync with README.md

Strategy: position-based section matching.
  - Parse H2 sections from all three READMEs as ordered lists.
  - Identify which EN sections changed by index (not heading text).
  - Check whether the same-index section in each translation also changed.
  - Warn if section counts diverge (structural drift).

This approach is robust to:
  - Heading renames in any language
  - Section reordering (treated as multiple changes)
  - New/removed sections (count mismatch warning)

Exit codes:
  1 - README.md changed but ZERO translations were touched (hard fail)
  0 - everything fine, or advisory warnings only (non-blocking)
"""

import hashlib
import os
import subprocess
import sys

FILES = {
    "en": "README.md",
    "zh": "README.zh-CN.md",
    "es": "README.es.md",
}


def run_git(*args):
    return subprocess.run(["git", *args], capture_output=True, text=True, encoding="utf-8")


def get_changed_files(base_sha):
    r = run_git("diff", "--name-only", base_sha, "HEAD")
    return {line.strip() for line in r.stdout.splitlines() if line.strip()}


def get_file_at_ref(path, ref):
    r = subprocess.run(["git", "show", f"{ref}:{path}"], capture_output=True)
    if r.returncode != 0:
        return None
    return r.stdout.decode("utf-8", errors="replace")


def parse_sections(text):
    """Return ordered list of (heading_line, body_text) for every H2 block."""
    sections = []
    heading = None
    body = []
    for line in text.splitlines(keepends=True):
        if line.startswith("## "):
            if heading is not None:
                sections.append((heading, "".join(body)))
            heading = line.rstrip("\n")
            body = []
        elif heading is not None:
            body.append(line)
    if heading is not None:
        sections.append((heading, "".join(body)))
    return sections


def sha(text):
    return hashlib.sha256(text.encode()).hexdigest()[:16]


def get_base_sha():
    base = os.environ.get("BASE_SHA", "").strip()
    if base:
        return base
    r = run_git("rev-parse", "HEAD~1")
    return r.stdout.strip() if r.returncode == 0 else None


def read_file(path):
    try:
        with open(path, encoding="utf-8") as f:
            return f.read()
    except FileNotFoundError:
        return None


def check_language(lang_key, en_curr_secs, en_prev_secs, changed_en_indices, base_sha, lang_changed):
    path = FILES[lang_key]
    curr_text = read_file(path)
    prev_text = get_file_at_ref(path, base_sha)
    issues = []

    if curr_text is None:
        issues.append(f"{path} not found in working tree.")
        return issues

    curr_secs = parse_sections(curr_text)
    prev_secs = parse_sections(prev_text) if prev_text else []

    # Structural drift check (advisory)
    if len(curr_secs) != len(en_curr_secs):
        issues.append(
            f"{path} has {len(curr_secs)} H2 section(s) vs README.md's {len(en_curr_secs)}. "
            "A section may have been added or removed without updating this translation."
        )

    if not lang_changed:
        issues.append(f"{path} was not updated even though README.md changed.")
        return issues

    # Per-section staleness check
    stale = []
    for i in sorted(changed_en_indices):
        if i >= len(curr_secs) or i >= len(prev_secs):
            continue
        if sha(curr_secs[i][1]) == sha(prev_secs[i][1]):
            en_heading = en_curr_secs[i][0] if i < len(en_curr_secs) else f"(section {i})"
            stale.append(f"    [{i}] EN: {en_heading!r}  =>  {path}: {curr_secs[i][0]!r}")

    if stale:
        issues.append(
            f"{path}: {len(stale)} section(s) look stale "
            f"(EN content changed but translation body unchanged):\n" + "\n".join(stale)
        )

    return issues


def main():
    base_sha = get_base_sha()
    if not base_sha:
        print("WARNING: cannot determine base SHA; skipping translation check.")
        sys.exit(0)

    changed = get_changed_files(base_sha)
    en_changed = FILES["en"] in changed
    zh_changed = FILES["zh"] in changed
    es_changed = FILES["es"] in changed

    if not en_changed:
        print("README.md unchanged; nothing to check.")
        sys.exit(0)

    # Hard fail: nobody updated any translation
    if not zh_changed and not es_changed:
        print("FAIL: README.md was updated but neither translation file was touched.")
        print(f"  Expected: {FILES['zh']}  and/or  {FILES['es']}")
        sys.exit(1)

    en_curr = read_file(FILES["en"])
    en_prev = get_file_at_ref(FILES["en"], base_sha)
    if not en_curr or not en_prev:
        print("WARNING: could not read README.md at both refs; skipping section check.")
        sys.exit(0)

    en_curr_secs = parse_sections(en_curr)
    en_prev_secs = parse_sections(en_prev)

    # Which EN sections (by position) actually changed?
    changed_en_indices = set()
    for i, (_, body) in enumerate(en_curr_secs):
        prev_body = en_prev_secs[i][1] if i < len(en_prev_secs) else ""
        if sha(body) != sha(prev_body):
            changed_en_indices.add(i)

    if not changed_en_indices:
        print("README.md changed but no H2 section bodies differ (e.g. only the language switcher or badges).")
        sys.exit(0)

    print(f"Changed section(s) in README.md ({len(changed_en_indices)}):")
    for i in sorted(changed_en_indices):
        h = en_curr_secs[i][0] if i < len(en_curr_secs) else f"(index {i})"
        print(f"  [{i}] {h}")

    all_issues = []
    for lang_key, lang_changed in [("zh", zh_changed), ("es", es_changed)]:
        issues = check_language(lang_key, en_curr_secs, en_prev_secs, changed_en_indices, base_sha, lang_changed)
        all_issues.extend(issues)

    if all_issues:
        print("\nADVISORY (not blocking - review and update translations if needed):")
        for issue in all_issues:
            print(f"  - {issue}")
    else:
        print("\nAll translation section checks passed.")

    sys.exit(0)


if __name__ == "__main__":
    main()
