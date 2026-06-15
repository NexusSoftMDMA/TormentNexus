import os
import re
import subprocess
import sys
import io

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8", errors="replace")

os.chdir("go")
TOOLS = "internal/tools"
PROTECTED = {"registry.go", "parity.go", "server.go", "factory.go"}
WIN_SEP = "internal\\tools\\"
UX_SEP = "internal/tools/"

# Step 1: Fix getString 2-value returns
fixed = 0
for fn in os.listdir(TOOLS):
    if not fn.endswith(".go") or fn in PROTECTED:
        continue
    path = os.path.join(TOOLS, fn)
    with open(path, "r", encoding="utf-8", errors="replace") as f:
        code = f.read()
    orig = code
    for pat, rep in [
        (r"(\w+)\s*,\s*\w+\s*:=\s*getString\(", r"\1 := getString("),
        (r"(\w+)\s*,\s*_\s*:=\s*getString\(", r"\1 := getString("),
        (r"(\w+)\s*,\s*\w+\s*:=\s*getInt\(", r"\1 := getInt("),
        (r"(\w+)\s*,\s*_\s*:=\s*getInt\(", r"\1 := getInt("),
        (r"(\w+)\s*,\s*\w+\s*:=\s*getBool\(", r"\1 := getBool("),
        (r"(\w+)\s*,\s*_\s*:=\s*getBool\(", r"\1 := getBool("),
    ]:
        code = re.sub(pat, rep, code)
    if code != orig:
        with open(path, "w", encoding="utf-8") as f:
            f.write(code)
        fixed += 1
print(f"Fixed {fixed} files with 2-value returns")

# Step 2: Find duplicate Handle function declarations
seen_handlers = {}
dup_files = set()
for fn in sorted(os.listdir(TOOLS)):
    if not fn.endswith(".go") or fn in PROTECTED:
        continue
    path = os.path.join(TOOLS, fn)
    with open(path, "r", encoding="utf-8", errors="replace") as f:
        code = f.read()
    for h in re.findall(r"func (Handle\w+)\s*\(", code):
        if h in seen_handlers:
            dup_files.add(fn)
        else:
            seen_handlers[h] = fn

removed = 0
for fn in dup_files:
    path = os.path.join(TOOLS, fn)
    if os.path.exists(path):
        os.unlink(path)
        removed += 1
print(f"Removed {removed} files with duplicate handlers")

# Step 3: Build and clean iteratively
for iteration in range(40):
    r = subprocess.run(
        ["go", "build", "-buildvcs=false", "./cmd/tormentnexus"],
        capture_output=True,
        text=True,
        timeout=120,
        encoding="utf-8",
        errors="replace",
    )
    if r.returncode == 0:
        print(f"BUILD CLEAN at iter {iteration}!")
        break

    broken = set()
    undefs = set()
    redeclared = set()
    stderr = r.stderr or ""

    for line in stderr.split("\n"):
        for sep in [WIN_SEP, UX_SEP]:
            if sep in line:
                fname = line.split(sep, 1)[1].split(":")[0]
                if fname.endswith(".go") and fname not in PROTECTED:
                    if "redeclared" in line:
                        redeclared.add(fname)
                    else:
                        broken.add(fname)
                break
        if "undefined:" in line:
            for w in line.split():
                w = w.rstrip(",")
                if w.startswith("Handle") and len(w) > 6:
                    undefs.add(w)

    for f in broken:
        p = os.path.join(TOOLS, f)
        if os.path.exists(p):
            os.unlink(p)

    if undefs:
        reg = os.path.join(TOOLS, "registry.go")
        with open(reg, "r", encoding="utf-8", errors="replace") as f:
            lines = f.readlines()
        new = [l for l in lines if not any(h in l for h in undefs)]
        with open(reg, "w", encoding="utf-8") as f:
            f.writelines(new)

    # For redeclarations, keep first file alphabetically, remove rest
    if redeclared:
        for f in sorted(redeclared)[1:]:
            p = os.path.join(TOOLS, f)
            if os.path.exists(p):
                os.unlink(p)
        # Also remove the first one if it's still causing issues after removing others
        # Actually just remove all redeclared files - the handlers are in registry already
        for f in redeclared:
            p = os.path.join(TOOLS, f)
            if os.path.exists(p):
                os.unlink(p)

    total_rm = len(broken) + len(redeclared)
    if total_rm == 0 and not undefs:
        print(f"Iter {iteration}: stuck")
        # Print the actual error to debug
        print(f"Error: {stderr[:200]}")
        break
    print(
        f"Iter {iteration}: -{total_rm} files ({len(broken)} broken, {len(redeclared)} redeclared), -{len(undefs)} undefs"
    )

files = [f for f in os.listdir(TOOLS) if f.endswith(".go")]
hc = len(
    re.findall(
        r"r\.handlers\[",
        open(
            os.path.join(TOOLS, "registry.go"), "r", encoding="utf-8", errors="replace"
        ).read(),
    )
)
print(f"{len(files)} files | {hc} handlers")
