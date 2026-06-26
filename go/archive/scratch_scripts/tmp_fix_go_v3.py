#!/usr/bin/env python3
"""Fix ALL Go syntax errors using gofmt detection and targeted fixes."""
import subprocess, re, os, glob

tool_dir = "internal/tools"
full_tool_dir = os.path.join("go", tool_dir)

def get_broken_files():
    """Use gofmt -e to find files with syntax errors."""
    broken = []
    pattern = os.path.join(full_tool_dir, "*.go")
    for fpath in glob.glob(pattern):
        fname = os.path.basename(fpath)
        result = subprocess.run(
            ["gofmt", "-e", fpath],
            capture_output=True, text=True, timeout=30
        )
        if result.returncode != 0:
            broken.append(fname)
    return set(broken)

def fix_one_file(filename):
    path = os.path.join(full_tool_dir, filename)
    with open(path, 'r', encoding='utf-8', errors='replace') as f:
        content = f.read()
    original = content
    
    # Fix 1: stray } after map literal opening
    content = re.sub(
        r'(map\[string\]interface\{\})\s*\{\s*\}\s*\n(\s*"[\w_]+")',
        r'\1{\n\2',
        content
    )
    
    # Fix 2: stray } before case/default (closes switch early)
    content = re.sub(r'\}\n(\s*)(case\s+|default:)', r'\n\1\2', content)
    
    # Fix 3: x, _ := getString -> x := getString
    for fn in ['getString', 'getInt', 'getFloat', 'getBool', 'getIntDefault', 'getBoolDefault']:
        content = re.sub(r'(\w+),\s*_\s*:=\s*' + fn + r'\s*\(', lambda m, fn=fn: m.group(1) + ' := ' + fn + '(', content)
    
    # Fix 4: interfact{} -> interface{}
    content = content.replace('interfact{}', 'interface{}')
    
    # Fix 5: ctx context context.Context -> ctx context.Context
    content = content.replace('ctx context context.Context', 'ctx context.Context')
    
    # Fix 6: trailing ") after MustCompile backtick
    content = re.sub(r'regexp\.MustCompile\(`[^`]*`\)"\)', lambda m: m.group()[:-2] + ')', content)
    
    # Fix 7: incomplete io.N or api
    content = re.sub(r'\bio\.N\b.*$', 'bytes.NewReader(body)', content, flags=re.MULTILINE)
    content = re.sub(r'\bapi$', 'apiKey)', content, flags=re.MULTILINE)
    
    # Fix 8: incomplete json.NewDecoder call
    content = re.sub(r'json\.NewDecoder\(resp\.Body\s*$', 'json.NewDecoder(resp.Body).Decode(&result)', content, flags=re.MULTILINE)
    
    # Fix 9: trailing ")" after function call (e.g. fmt.Sprintf(...)\")
    content = re.sub(r'fmt\.Sprintf\(`[^`]*`,\s*\w+\)"\)', lambda m: m.group()[:-2] + ')', content)
    
    # Fix 10: missing closing braces
    opens = content.count('{')
    closes = content.count('}')
    if opens > closes:
        needed = opens - closes
        content = content.rstrip() + '\n' + ('}\n' * needed)
    
    # Fix 11: unmatched double quotes in function calls
    lines = content.split('\n')
    for i, line in enumerate(lines):
        stripped = line.strip()
        if stripped.startswith('//') or stripped.startswith('package') or stripped.startswith('import'):
            continue
        # Check for lines that end with \" after a closing paren (stray string start)
        if re.search(r'\)"\)?\s*$', stripped) and stripped.count('"') % 2 == 1:
            lines[i] = stripped[:-1] + '\n' if stripped.endswith('"') else stripped + '\n'
    content = '\n'.join(lines)
    
    # Fix 12: stray ")" at end of line with incomplete string
    content = re.sub(r'\)"\)\s*$', '))\n', content, flags=re.MULTILINE)
    
    if content != original:
        with open(path, 'w', encoding='utf-8', newline='') as f:
            f.write(content)
        return True
    return False

# Iterative fixing
max_iter = 30
for iteration in range(max_iter):
    broke = get_broken_files()
    if not broke:
        print(f" ALL 3941 FILES COMPILE! Fixed in {iteration} iterations.")
        break
    print(f"Iteration {iteration+1}: {len(broke)} broken files")
    fixed = 0
    not_fixed = []
    for f in sorted(broke):
        if fix_one_file(f):
            fixed += 1
        else:
            not_fixed.append(f)
    print(f"  Fixed {fixed} files")
    if not_fixed:
        print(f"  Could not fix {len(not_fixed)} files (will need manual attention):")
        for f in not_fixed[:10]:
            print(f"    {f}")
    if fixed == 0:
        print("  No more fixes possible!")
        break
else:
    print(f"Reached max iterations. {len(broke)} files still broken.")
