#!/usr/bin/env python3
"""
Add //go:build ignore tags to all Go files in go/internal/tools
to exclude them from the default build.
"""

import pathlib

ROOT = pathlib.Path("go/internal/tools")

TAG_LINES = [
    "//go:build ignore",
    "// +build ignore",
]


def add_ignore_tag(file_path: pathlib.Path):
    content = file_path.read_text(encoding="utf-8", errors="replace")
    if any(tag in content for tag in TAG_LINES):
        return  # already tagged
    lines = content.splitlines()
    insert_idx = 0
    for i, line in enumerate(lines):
        stripped = line.strip()
        if stripped == "" or stripped.startswith("//"):
            continue
        insert_idx = i
        break
    new_lines = lines[:insert_idx] + TAG_LINES + [""] + lines[insert_idx:]
    file_path.write_text("\n".join(new_lines) + "\n", encoding="utf-8")


def main():
    count = 0
    for go_file in ROOT.glob("*.go"):
        add_ignore_tag(go_file)
        count += 1
    print(f"Added ignore tags to {count} files.")


if __name__ == "__main__":
    main()
