#!/usr/bin/env python3
"""Render a release-notes Markdown file into a small standalone HTML
fragment suitable for inlining into a Sparkle appcast `<description>`
element (CDATA-wrapped).

Stdlib only — no `markdown` package dependency. Covers the subset of
GitHub-flavored markdown that `releases/v*/RELEASE_NOTES.md` uses:

* `## Heading 2` / `### Heading 3`
* paragraphs (blank-line-separated)
* unordered lists (`- item`, single level only)
* fenced code blocks (` ``` `)
* inline `code`, **bold**, *italic*, `[link text](url)`
* horizontal rules (`---`)

Sparkle's `SUUserUpdateAlertController` renders the inline HTML in a
WebKit view with no styling beyond what's in the body, so a tiny
`<style>` block is included. Fonts and spacing are tuned to look
right inside the standard 480×360 update sheet.

Usage:
    python3 tools/render-release-notes.py releases/v2.7.0/RELEASE_NOTES.md > out.html

Used by `scripts/release.sh` to populate the appcast item's
`<description>` block per release.
"""
from __future__ import annotations
import html
import re
import sys
from pathlib import Path
from typing import Iterator


# ---------- inline ----------

_INLINE_CODE = re.compile(r"`([^`]+)`")
_BOLD = re.compile(r"\*\*([^*]+)\*\*")
_ITALIC = re.compile(r"(?<!\*)\*([^*]+)\*(?!\*)")
_LINK = re.compile(r"\[([^\]]+)\]\(([^)]+)\)")


def render_inline(text: str) -> str:
    """Apply inline transforms in order: escape HTML first, then
    swap markdown markers in-place. Order matters — links before
    bold so `[**bold**](url)` doesn't double-process."""
    out = html.escape(text)
    out = _INLINE_CODE.sub(lambda m: f"<code>{m.group(1)}</code>", out)
    out = _LINK.sub(lambda m: f'<a href="{m.group(2)}">{m.group(1)}</a>', out)
    out = _BOLD.sub(lambda m: f"<strong>{m.group(1)}</strong>", out)
    out = _ITALIC.sub(lambda m: f"<em>{m.group(1)}</em>", out)
    return out


# ---------- block ----------

def render_blocks(lines: list[str]) -> Iterator[str]:
    """Walk lines and emit HTML blocks. Maintains state for fenced
    code, lists, and paragraph buffers."""
    i = 0
    n = len(lines)
    paragraph_buf: list[str] = []
    list_buf: list[str] = []

    def flush_paragraph() -> Iterator[str]:
        if paragraph_buf:
            text = " ".join(paragraph_buf).strip()
            if text:
                yield f"<p>{render_inline(text)}</p>"
            paragraph_buf.clear()

    def flush_list() -> Iterator[str]:
        if list_buf:
            yield "<ul>"
            for item in list_buf:
                yield f"  <li>{render_inline(item)}</li>"
            yield "</ul>"
            list_buf.clear()

    while i < n:
        line = lines[i]
        stripped = line.rstrip("\n")

        # Fenced code block
        if stripped.startswith("```"):
            yield from flush_paragraph()
            yield from flush_list()
            i += 1
            code_lines: list[str] = []
            while i < n and not lines[i].rstrip("\n").startswith("```"):
                code_lines.append(lines[i].rstrip("\n"))
                i += 1
            i += 1  # skip closing fence
            escaped = html.escape("\n".join(code_lines))
            yield f"<pre><code>{escaped}</code></pre>"
            continue

        # Blank line — close paragraph + list
        if not stripped.strip():
            yield from flush_paragraph()
            yield from flush_list()
            i += 1
            continue

        # Horizontal rule
        if stripped.strip() == "---":
            yield from flush_paragraph()
            yield from flush_list()
            yield "<hr>"
            i += 1
            continue

        # Heading
        if stripped.startswith("### "):
            yield from flush_paragraph()
            yield from flush_list()
            yield f"<h3>{render_inline(stripped[4:])}</h3>"
            i += 1
            continue
        if stripped.startswith("## "):
            yield from flush_paragraph()
            yield from flush_list()
            yield f"<h2>{render_inline(stripped[3:])}</h2>"
            i += 1
            continue
        if stripped.startswith("#### "):
            yield from flush_paragraph()
            yield from flush_list()
            yield f"<h4>{render_inline(stripped[5:])}</h4>"
            i += 1
            continue

        # Unordered list item
        list_match = re.match(r"^[-*]\s+(.+)$", stripped)
        if list_match:
            yield from flush_paragraph()
            list_buf.append(list_match.group(1))
            i += 1
            continue

        # Paragraph line — close list, accumulate
        if list_buf:
            yield from flush_list()
        paragraph_buf.append(stripped)
        i += 1

    yield from flush_paragraph()
    yield from flush_list()


# ---------- document ----------

# Sparkle WebKit view default styling is plain — give it enough to
# look like a release notes sheet, not a 1995 docs dump. Sized for
# the standard update alert dimensions.
STYLE = """\
body {
  font-family: -apple-system, BlinkMacSystemFont, "Helvetica Neue", sans-serif;
  font-size: 13px;
  line-height: 1.5;
  color: #1d1d1f;
  margin: 0;
  padding: 0 4px;
}
h2 {
  font-size: 17px;
  margin: 16px 0 6px 0;
  border-bottom: 1px solid #e5e5e7;
  padding-bottom: 3px;
}
h3 {
  font-size: 14px;
  margin: 14px 0 4px 0;
  color: #424245;
}
h4 {
  font-size: 13px;
  font-weight: 600;
  margin: 10px 0 2px 0;
}
p { margin: 6px 0; }
ul { margin: 6px 0; padding-left: 20px; }
li { margin: 3px 0; }
code {
  background: #f5f5f7;
  border-radius: 3px;
  padding: 1px 4px;
  font-family: "SF Mono", Menlo, Consolas, monospace;
  font-size: 12px;
}
pre {
  background: #f5f5f7;
  border-radius: 5px;
  padding: 8px 10px;
  overflow-x: auto;
  font-size: 12px;
}
pre code { background: transparent; padding: 0; }
a { color: #0066cc; text-decoration: none; }
a:hover { text-decoration: underline; }
hr {
  border: none;
  border-top: 1px solid #e5e5e7;
  margin: 16px 0;
}
strong { color: #1d1d1f; }
@media (prefers-color-scheme: dark) {
  body { color: #f5f5f7; background: #1c1c1e; }
  h2 { border-bottom-color: #38383a; }
  h3 { color: #c7c7cc; }
  code, pre { background: #2c2c2e; }
  hr { border-top-color: #38383a; }
  a { color: #4499ff; }
  strong { color: #f5f5f7; }
}
"""


def render_document(markdown: str) -> str:
    body = "\n".join(render_blocks(markdown.splitlines(keepends=True)))
    return f"<!DOCTYPE html><html><head><meta charset=\"utf-8\"><style>{STYLE}</style></head><body>\n{body}\n</body></html>"


def main(argv: list[str]) -> int:
    if len(argv) != 2:
        sys.stderr.write("usage: render-release-notes.py <RELEASE_NOTES.md>\n")
        return 2
    path = Path(argv[1])
    if not path.exists():
        sys.stderr.write(f"file not found: {path}\n")
        return 1
    sys.stdout.write(render_document(path.read_text(encoding="utf-8")))
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
