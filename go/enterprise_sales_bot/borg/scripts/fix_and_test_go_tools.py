#!/usr/bin/env python3
"""
Comprehensive fixer for broken Go tools files.
Handles ALL known error patterns identified from gofmt -e.
"""
import pathlib, re, subprocess, sys, shutil

ROOT = pathlib.Path("go/internal/tools")
BROKEN = ROOT / "current_broken.txt"
STILL_BROKEN = ROOT / "go_tools_still_broken.txt"


def gofmt_ok(p: pathlib.Path) -> bool:
    r = subprocess.run(
        ["gofmt", "-e", str(p)],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    return r.returncode == 0


def fix_content(txt: str) -> str:
    # 1. Remove corruption markers: --- lines and model reference lines
    txt = re.sub(r'^---\s*$', '', txt, flags=re.MULTILINE)
    txt = re.sub(r'^\*\[?(Model|Provider):.*$', '', txt, flags=re.MULTILINE)
    txt = re.sub(r'^\*?(deepseek-reasoner|openai|qwen|mistral|nvidia|huggingface).*$', '', txt, flags=re.MULTILINE | re.IGNORECASE)
    # Remove any remaining model reference fragments
    txt = re.sub(r'^.*-reasoner\s*\(.*\)\s*,?\s*$', '', txt, flags=re.MULTILINE)

    # 2. Fix trailing commas in imports: `"net/http",` -> `"net/http"`
    txt = re.sub(r'^(\s*"\w[\w/._-]*"),\s*$', r'\1', txt, flags=re.MULTILINE)

    # 3. Fix trailing commas after return: `return err("..."),` -> `return err("...")`
    txt = re.sub(r'(return\s+[^;{]+?),(\s*)$', r'\1\2', txt, flags=re.MULTILINE)

    # 4. Fix trailing commas in assignments: `x = "value",` -> `x = "value"`
    #    and `body["key"] = value,` -> `body["key"] = value`
    txt = re.sub(r'^(\s*\S+?\s*=\s*[^,;{]+?),(\s*)$', r'', txt, flags=re.MULTILINE)

    # 5. Fix stray comma after opening map brace: `{,` -> `{`
    txt = re.sub(r'\{\s*,', '{', txt)

    # 6. Fix stray `")` at end of function call lines
    #    Remove trailing `")` or just `"` that shouldn't be there
    txt = re.sub(r'\)"\)\s*$', ')', txt, flags=re.MULTILINE)   # ...))") -> ...))
    txt = re.sub(r'\)"\s*$', ')', txt, flags=re.MULTILINE)     # ...))" -> ...))

    # 7. Remove isolated `}` lines that sit before a quoted map key (stray closing brace)
    lines = txt.splitlines()
    cleaned = []
    i = 0
    while i < len(lines):
        line = lines[i]
        stripped = line.strip()
        # Check if this is a stray `}` followed by a map key line
        if stripped == '}':
            nxt = lines[i + 1].strip() if i + 1 < len(lines) else ''
            if nxt.startswith('"') and (':' in nxt or ',' in nxt):
                i += 1  # skip stray }
                continue
        cleaned.append(line)
        i += 1
    txt = '\n'.join(cleaned)

    # 8. Remove stray `}` before keywords: `}\nif` or `}\nswitch` or `}\nreturn`
    #    This happens when generator adds an extra closing brace
    txt = re.sub(r'\}\s*\n(\s*)(if|switch|return|for)\b', r'\1\2', txt)

    # 9. Fix missing parentheses in function signatures: `func HandleXyz ctx context.Context,...`
    #    -> `func HandleXyz(ctx context.Context, args map[string]interface{}) (ToolResponse, error)`
    txt = re.sub(
        r'func (\w+)\s+([a-zA-Z_]\w*)\s+([a-zA-Z_]\w*)\.([a-zA-Z_]\w*),?\s+([a-zA-Z_]\w*)\s+map\[string\]interface\{\}\s*\{',
        r'func \1(\2 \3.\4, \5 map[string]interface{}) (ToolResponse, error) {',
        txt
    )

    # 10. Add missing comma before newline in function arguments
    #     Pattern: argument followed by newline then ) -> add comma
    txt = re.sub(r'(\w[\w.]*)\s*\n(\s*)\)', r'\1,\n\2)', txt)

    # 11. Add missing comma before newline in composite literals
    #     Pattern: value followed by newline then } -> add comma
    txt = re.sub(r'("[^"]*"|[\w.]+)\s*\n(\s*)\}', r'\1,\n\2}', txt)

    # 12. Balance braces - append missing closing braces
    opens = txt.count('{')
    closes = txt.count('}')
    needed = max(0, opens - closes)
    if needed:
        # But don't over-append - try up to 3 braces max
        limit = min(needed, 5)
        txt = txt.rstrip() + ('\n}' * limit) + '\n'

    return txt


def process_file(fpath: pathlib.Path) -> bool:
    orig = fpath.read_text(encoding='utf-8', errors='replace')
    mod = fix_content(orig)
    if mod != orig:
        fpath.write_text(mod, encoding='utf-8', newline='\n')
    return gofmt_ok(fpath)


def main():
    if not BROKEN.is_file():
        alt = pathlib.Path('current_broken.txt')
        if alt.is_file():
            shutil.copy(alt, BROKEN)
        else:
            print(f'ERROR: {BROKEN} not found')
            sys.exit(1)

    broken_files = [
        ROOT / p for p in BROKEN.read_text().splitlines() if p.strip()
    ]
    total = len(broken_files)
    success = 0
    still = []

    for idx, f in enumerate(broken_files, 1):
        if not f.exists():
            print(f"[{idx}/{total}] {f.name}: MISSING")
            still.append(f.name)
            continue
        print(f"[{idx}/{total}] {f.name} ...", end=' ', flush=True)
        if process_file(f):
            success += 1
            print('OK')
        else:
            still.append(f.name)
            print('FAIL')

    print(f'\n=== SUMMARY ===')
    print(f'Processed: {total}')
    print(f'Fixed:      {success}')
    print(f'Still broken: {len(still)}')

    if still:
        STILL_BROKEN.write_text('\n'.join(still) + '\n', encoding='utf-8')

    print('\nRunning go build...')
    b = subprocess.run(
        ['go', 'build', './internal/tools/...'],
        cwd=pathlib.Path('go'),
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    out = b.stdout + '\n' + b.stderr
    lines = out.strip().split('\n')
    # Show first 20 errors
    for line in lines[:20]:
        print(line)


if __name__ == '__main__':
    main()
