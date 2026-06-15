#!/usr/bin/env python3
"""Fix ALL Go syntax errors in the tools directory by detecting with gofmt and applying known fixes."""
import subprocess, re, os, glob

tool_dir = "go/internal/tools"

def get_broken_files():
    """Use gofmt -e to find files with syntax errors."""
    result = subprocess.run(
        ["find", "internal/tools", "-name", "*.go", "-exec", "sh", "-c", 
         'gofmt -e "$1" > /dev/null 2>&1 || echo "$1"', "_", "{}", ";"],
        cwd="go", capture_output=True, text=True, timeout=120
    )
    files = set()
    for line in result.stdout.strip().split('\n'):
        line = line.strip()
        if line:
            files.add(os.path.basename(line))
    return files

def fix_one_file(filename):
    path = os.path.join(tool_dir, filename)
    with open(path, 'r', encoding='utf-8', errors='replace') as f:
        content = f.read()
    original = content
    
    # Fix 1: stray } after map literal opening (most common error)
    # Pattern: return TYPE{\n}\n"key" -> return TYPE{\n"key"
    content = re.sub(
        r'(return\s+\w+\(?\s*map\[string\]interface\{\}\{)',
        lambda m: m.group(1),
        content
    )
    content = re.sub(
        r'(map\[string\]interface\{\})\s*\{\s*\}\s*\n(\s*"[\w_]+")',
        r'\1{\n\2',
        content
    )
    
    # Fix 2: stray } before case/default
    content = re.sub(r'\}\n(\s*)(case|default)\s+', r'\n\1\2 ', content)
    
    # Fix 3: x, _ := getString -> x := getString (and similar)
    for fn in ['getString', 'getInt', 'getFloat', 'getBool', 'getIntDefault', 'getBoolDefault']:
        content = re.sub(r'(\w+),\s*_\s*:=\s*' + fn + r'\s*\(', r'\1 := \2(', content)
    
    # Fix 4: interfact{} -> interface{}
    content = content.replace('interfact{}', 'interface{}')
    
    # Fix 5: ctx context context.Context -> ctx context.Context
    content = content.replace('ctx context context.Context', 'ctx context.Context')
    
    # Fix 6: incomplete concatenation at end of line
    for func in ['ok', 'err']:
        content = re.sub(
            r'return ' + func + r'\("(?:[^"]*)"\s*\+\s*$',
            lambda m: m.group() + ' string(body))',
            content,
            flags=re.MULTILINE
        )
    
    # Fix 7: trailing \") after MustCompile backtick
    content = re.sub(r'regexp\.MustCompile\(`[^`]*`\)"\)', lambda m: m.group()[:-2] + ')', content)
    
    # Fix 8: truncated io.N or api
    content = re.sub(r'\bio\.N\b.*$', 'bytes.NewReader(body)', content, flags=re.MULTILINE)
    content = re.sub(r'\bapi$', 'apiKey)', content, flags=re.MULTILINE)
    
    # Fix 9: incomplete json.NewDecoder call
    content = re.sub(r'json\.NewDecoder\(resp\.Body\s*$', 'json.NewDecoder(resp.Body).Decode(&result)', content, flags=re.MULTILINE)
    
    # Fix 10: missing closing braces (add if needed)
    opens = content.count('{')
    closes = content.count('}')
    if opens > closes:
        needed = opens - closes
        content = content.rstrip() + '\n' + ('}\n' * needed)
    
    # Fix 11: unterminated strings - find lines with odd number of quotes
    lines = content.split('\n')
    for i, line in enumerate(lines):
        stripped = line.strip()
        # Skip comments and package/import lines
        if stripped.startswith('//') or stripped.startswith('package') or stripped.startswith('import'):
            continue
        # Count double quotes (excluding those in backtick strings)
        in_backtick = False
        quote_count = 0
        for ch in stripped:
            if ch == '`':
                in_backtick = not in_backtick
            elif ch == '"' and not in_backtick:
                quote_count += 1
        if quote_count % 2 == 1:
            # Odd number of quotes - close with \"
            lines[i] = line.rstrip() + '"\n'
    content = '\n'.join(lines)
    
    if content != original:
        with open(path, 'w', encoding='utf-8', newline='') as f:
            f.write(content)
        return True
    return False

# Iterative fixing
max_iter = 20
for iteration in range(max_iter):
    broke = get_broken_files()
    if not broke:
        print(f"ALL FIXED after {iteration} iterations!")
        break
    print(f"Iteration {iteration+1}: {len(broke)} broken files")
    fixed = 0
    for f in broke:
        if fix_one_file(f):
            fixed += 1
    print(f"  Fixed {fixed} files")
    if fixed == 0:
        print("  No more fixes possible. Remaining:")
        for f in sorted(broke)[:10]:
            print(f"    {f}")
        break
else:
    print(f"Reached max iterations. {len(broke)} files still broken.")
    for f in sorted(broke):
        print(f"  {f}")
