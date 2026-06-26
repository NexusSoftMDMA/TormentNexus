#!/usr/bin/env python3
import re
import os


def fix_file(fpath):
    with open(fpath, "r", encoding="utf-8", errors="replace") as f:
        content = f.read()
    orig = content

    # Fix stray } before case/default
    content = re.sub(r"\}\n(\s*)(case\s+|default:)", r"\n\1\2", content)

    # Fix stray } after map{}
    content = re.sub(r"(map\[string\]interface\{\})\s*\{\s*\}\s*\n", r"\1{\n", content)

    # Fix x, _ := getXX
    for fn in ["getString", "getInt", "getFloat", "getBool"]:
        content = re.sub(rf"(\w+),\s*_\s*:=\s*{fn}\s*\(", rf"\1 := {fn}(", content)

    # Fix interfact
    content = content.replace("interfact{}", "interface{}")

    # Fix ctx context context.Context
    content = content.replace("ctx context context.Context", "ctx context.Context")

    # Fix trailing ") after MustCompile backtick
    content = re.sub(
        r'regexp\.MustCompile\(`[^`]*`\)"\)', lambda m: m.group()[:-2] + ")", content
    )

    if content != orig:
        with open(fpath, "w", encoding="utf-8", newline="") as f:
            f.write(content)
        return True
    return False


with open("broken_list.txt", "r") as f:
    files = [l.strip() for l in f if l.strip()]

print(f"Processing {len(files)} broken files...")
fixed = 0
for fname in files:
    fpath = os.path.join("go/internal/tools", fname)
    if os.path.exists(fpath):
        if fix_file(fpath):
            fixed += 1
            print(f"Fixed {fname}")

print(f"Done. Fixed {fixed} files.")
