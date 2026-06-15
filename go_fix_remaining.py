#!/usr/bin/env python3
import re
import os
import subprocess
import sys

TOOLS = "go/internal/tools"
BROKEN_FILE = "broken_list2.txt"


def get_broken():
    with open(BROKEN_FILE) as f:
        return [l.strip() for l in f if l.strip()]


def fixes_for_content(content):
    # Fix 1: stray }") after map[string]interface{}{ closing brace
    content = re.sub(r"(map\[string\]interface\{\})\s*\{\s*\}\s*\n", r"\1{\n", content)

    # Fix 2: stray }") before case/default
    content = re.sub(r"\}\n\s*(case\s+|default:)", r"\n\1", content)

    # Fix 3: interfact{} -> interface{}
    content = content.replace("interfact{}", "interface{}")

    # Fix 4: ctx context context.Context -> ctx context.Context
    content = content.replace("ctx context context.Context", "ctx context.Context")

    # Fix 5: x, _ := getString -> x := getString (for common getters)
    for fn in ["getString", "getInt", "getFloat", "getBool"]:
        content = re.sub(rf"(\w+),\s*_\s*:=\s*{fn}\s*\(", rf"\1 := {fn}(", content)

    # Fix 6: trailing ")" after MustCompile backtick string call
    content = re.sub(
        r'regexp\.MustCompile\(`[^`]*`\)"\)', lambda m: m.group()[:-2] + ")", content
    )

    # Fix 7: remove stray " before case
    content = re.sub(r'"case\b', "case", content)

    # Fix 8: unterminated strings: if line has odd quotes, add closing quote
    lines = content.split("\n")
    changed_lines = False
    for i in range(len(lines)):
        sl = lines[i]
        if any(
            sl.startswith(p)
            for p in ("//", "package ", "import ", "func ", "type ", "const ", "var ")
        ):
            continue
        if (
            sl.strip().startswith("case ")
            or sl.strip().startswith("default:")
            or "\ncase " in sl
        ):
            continue
        # Count open double quotes not inside backticks
        in_bt = False
        qopen = []
        for idx, ch in enumerate(sl):
            if ch == "`":
                in_bt = not in_bt
            elif ch == '"' and not in_bt:
                qopen.append(idx)
        if len(qopen) % 2 == 1:
            # Unterminated
            if sl.rstrip().endswith('")'):
                lines[i] = sl.rstrip()[:-2] + '")' + "\n"
            else:
                lines[i] = sl.rstrip() + '"\n'
            changed_lines = True
    if changed_lines:
        content = "\n".join(lines)

    return content


fixed = []
again = set()

for iteration in range(10):
    broken = get_broken()
    if not broken:
        print("ALL FIXED!")
        sys.exit(0)
    print(f"Iteration {iteration + 1}: {len(broken)} broken")

    for fname in sorted(broken):
        fpath = os.path.join(TOOLS, fname)
        if not os.path.exists(fpath):
            again.add(fname)
            continue
        with open(fpath, "r", encoding="utf-8", errors="replace") as f:
            orig = f.read()
        mod = fixes_for_content(orig)
        if mod != orig:
            with open(fpath, "w", encoding="utf-8", newline="") as f:
                f.write(mod)
            print(f"Fixed {fname}")
            fixed.append(fname)
        # re-validate: if gofmt now accepts the file, remove from broken list
        r = subprocess.run(
            ["gofmt", "-e", fpath], capture_output=True, text=True, timeout=30
        )
        if r.returncode == 0:
            fixed.append(fname)
        else:
            # file still broken; leave it in broken list for next iteration
            pass

    # If no fixes applied and no files finished, break
    if not fixed:
        print("Could not fix any more files.")
        for f in sorted(broken)[:10]:
            print(f"  {f}")
        break
    fixed.clear()

print("Remaining broken:")
with open(BROKEN_FILE, "r") as f:
    for line in f:
        if line.strip() in again:
            print(f" MISSING: {line.strip()}")
        else:
            print(line.strip())
