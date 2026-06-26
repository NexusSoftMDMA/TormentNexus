"""``jcodemunch-mcp whatsnew`` — auto-generate the README recency block + whatsnew.json.

Reads the top N release entries from ``CHANGELOG.md`` and writes:

1. ``<repo-root>/whatsnew.json`` — machine-readable artifact published as
   a release asset; also fetched at first launch by the version-drift
   probe in :mod:`jcodemunch_mcp.version_check`.
2. The README block between ``<!-- WHATSNEW:START -->`` and
   ``<!-- WHATSNEW:END -->`` — keeps the top-of-fold "What's new" panel
   fresh on every release without hand-editing.

Run as part of the release flow before ``python -m build``.
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path

_CHANGELOG_HEADER_RE = re.compile(
    r"^##\s*\[(?P<version>[\d.]+)\]\s*[-—]\s*(?P<date>\d{4}-\d{2}-\d{2})\s*[-—]\s*(?P<title>.+)$"
)

_WHATSNEW_BEGIN = "<!-- WHATSNEW:START -->"
_WHATSNEW_END = "<!-- WHATSNEW:END -->"


def parse_changelog(path: Path, max_entries: int = 3) -> list[dict]:
    """Return the top `max_entries` release entries from CHANGELOG.md.

    Each entry: {version, date, title, summary} where summary is the
    first prose paragraph after the heading (or empty if none).
    """
    if not path.exists():
        return []

    lines = path.read_text(encoding="utf-8").splitlines()
    entries: list[dict] = []
    i = 0
    while i < len(lines) and len(entries) < max_entries:
        m = _CHANGELOG_HEADER_RE.match(lines[i])
        if not m:
            i += 1
            continue
        version, date, title = m["version"], m["date"], m["title"].strip()

        # Walk forward to capture the first prose paragraph (skip blanks +
        # subheadings until we find content; stop at the next release header).
        summary_lines: list[str] = []
        j = i + 1
        while j < len(lines):
            line = lines[j]
            if _CHANGELOG_HEADER_RE.match(line):
                break
            stripped = line.strip()
            if not stripped:
                if summary_lines:
                    break  # paragraph ended
                j += 1
                continue
            if stripped.startswith("###"):
                # Skip subheadings unless we haven't found content yet
                j += 1
                continue
            summary_lines.append(stripped)
            j += 1
            if len(summary_lines) >= 4:
                break

        entries.append({
            "version": version,
            "date": date,
            "title": title,
            "summary": " ".join(summary_lines).strip(),
        })
        i = j

    return entries


def render_readme_block(entries: list[dict], repo: str) -> str:
    """Render the README "What's new" block as markdown."""
    if not entries:
        return ""
    lines = ["#### What's new", ""]
    for e in entries:
        link = f"https://github.com/{repo}/releases/tag/v{e['version']}"
        lines.append(f"- **[v{e['version']}]({link})** ({e['date']}) — {e['title']}")
    return "\n".join(lines)


def update_readme(readme_path: Path, block: str) -> bool:
    """Replace content between WHATSNEW markers in README. Returns True if changed."""
    if not readme_path.exists():
        return False
    text = readme_path.read_text(encoding="utf-8")

    if _WHATSNEW_BEGIN not in text or _WHATSNEW_END not in text:
        return False

    pattern = re.compile(
        re.escape(_WHATSNEW_BEGIN) + r"(.*?)" + re.escape(_WHATSNEW_END),
        re.DOTALL,
    )
    replacement = f"{_WHATSNEW_BEGIN}\n{block}\n{_WHATSNEW_END}"
    new_text = pattern.sub(replacement, text)
    if new_text == text:
        return False
    readme_path.write_text(new_text, encoding="utf-8")
    return True


def write_whatsnew_json(
    out_path: Path, entries: list[dict], current_version: str
) -> None:
    """Write whatsnew.json artifact for release asset + drift probe."""
    payload = {
        "current": current_version,
        "entries": entries,
    }
    out_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")


def _read_version(pyproject_path: Path) -> str:
    """Extract version = "X.Y.Z" from pyproject.toml."""
    text = pyproject_path.read_text(encoding="utf-8")
    m = re.search(r'^version\s*=\s*"([^"]+)"', text, re.MULTILINE)
    if not m:
        raise RuntimeError(f"version not found in {pyproject_path}")
    return m.group(1)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(
        description="Generate whatsnew.json and refresh README recency block."
    )
    parser.add_argument(
        "--repo-root",
        type=Path,
        default=Path.cwd(),
        help="Repository root (default: cwd)",
    )
    parser.add_argument(
        "--repo",
        default="jgravelle/jcodemunch-mcp",
        help="GitHub repo slug for release links",
    )
    parser.add_argument(
        "--max-entries",
        type=int,
        default=3,
        help="Number of recent releases to include",
    )
    args = parser.parse_args(argv)

    root: Path = args.repo_root
    changelog = root / "CHANGELOG.md"
    readme = root / "README.md"
    whatsnew = root / "whatsnew.json"
    pyproject = root / "pyproject.toml"

    entries = parse_changelog(changelog, max_entries=args.max_entries)
    if not entries:
        print(f"warning: no entries parsed from {changelog}", file=sys.stderr)
        return 1

    current_version = _read_version(pyproject)

    write_whatsnew_json(whatsnew, entries, current_version)
    print(f"wrote {whatsnew} ({len(entries)} entries, current={current_version})")

    block = render_readme_block(entries, args.repo)
    if update_readme(readme, block):
        print(f"updated {readme} WHATSNEW block")
    else:
        print(
            f"note: {readme} not updated — markers missing or block already current",
            file=sys.stderr,
        )

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
