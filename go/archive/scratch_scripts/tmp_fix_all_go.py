#!/usr/bin/env python3
"""Iterative Go syntax fixer - runs build, collects errors, fixes files, repeats."""
import subprocess, re, os, sys

tool_dir = "go/internal/tools"

def get_errors():
    result = subprocess.run(
        ["go", "build", "./internal/tools/..."],
        cwd="go",
        capture_output=True, text=True, timeout=120
    )
    return result.stderr

def parse_errors(stderr):
    """Extract unique (file, line) pairs from build errors."""
    files = set()
    for line in stderr.split('\n'):
        # Match pattern: internal\tools\filename.go:LINE:COL: error
        m = re.match(r'internal\\tools\\([^:]+):(\d+):', line)
        if m:
            files.add((m.group(1), int(m.group(2))))
    return files

def fix_file(filename):
    path = os.path.join(tool_dir, filename)
    with open(path, 'r', encoding='utf-8', errors='replace') as f:
        content = f.read()
    
    original = content
    changes = []
    
    # Pattern 1: x, _ := getString(...) -> x := getString(...)
    # Also handle getInt, getFloat, getBool, getBoolDefault etc.
    content = re.sub(
        r'(\w+),\s*_\s*:=\s*(getString|getInt|getFloat|getBool|getBoolDefault|getIntDefault)',
        r'\1 := \2',
        content
    )
    
    # Pattern 2: stray } before case/default (closes switch early)
    content = re.sub(
        r'\}\n(\t*)case ', 
        r'\n\1case ', 
        content
    )
    content = re.sub(
        r'\}\n(\t*)default:', 
        r'\n\1default:', 
        content
    )
    
    # Pattern 3: interfact{} -> interface{}
    content = content.replace('interfact{}', 'interface{}')
    
    # Pattern 4: ctx context context.Context -> ctx context.Context
    content = content.replace('ctx context context.Context', 'ctx context.Context')
    
    # Pattern 5: incomplete return ok/err with unterminated concatenation
    # return ok("..." + \n} -> complete with string(body)
    content = re.sub(
        r'return ok\(("[^"]*")\s*\+\s*$',
        r'return ok(\1 + string(body))',
        content,
        flags=re.MULTILINE
    )
    content = re.sub(
        r'return err\(("[^"]*")\s*\+\s*$',
        r'return err(\1 + string(body))',
        content,
        flags=re.MULTILINE
    )
    
    # Pattern 6: incomplete io.N or similar truncated expressions
    content = re.sub(
        r'io\.N\b.*$',
        'bytes.NewReader(body)',
        content,
        flags=re.MULTILINE
    )
    
    # Pattern 7: incomplete json.NewDecoder(resp.Body or similar
    content = re.sub(
        r'json\.NewDecoder\(resp\.Body\s*$',
        'json.NewDecoder(resp.Body).Decode(&result)',
        content,
        flags=re.MULTILINE
    )
    
    # Pattern 8: trailing ")" at end of some function calls
    content = re.sub(
        r'strings\.ReplaceAll\(text,\s*`"`,\s*`\\"`\)\)"\)',
        'strings.ReplaceAll(text, `"`, `\\"`))',
        content
    )
    
    # Pattern 9: missing closing brace for function after switch
    # Count braces
    opens = content.count('{')
    closes = content.count('}')
    if opens > closes:
        content = content.rstrip() + '\n}\n'
    
    # Remove model reference markers (--- and model name lines)
    lines = content.split('\n')
    new_lines = []
    skip = False
    for i, line in enumerate(lines):
        if line.strip() == '---':
            # Check if next line is model reference
            if i + 1 < len(lines):
                next_line = lines[i + 1].strip()
                if any(p in next_line for p in ['deepseek', 'openai/', 'qwen', 'Mistral', 'huggingface', 'nvidia']):
                    skip = True
                    continue
        if skip:
            skip = False
            continue
        new_lines.append(line)
    content = '\n'.join(new_lines)
    
    if content != original:
        with open(path, 'w', encoding='utf-8', newline='') as f:
            f.write(content)
        return True
    return False

# Iterative fixing
max_iterations = 10
for iteration in range(max_iterations):
    print(f"\n=== Iteration {iteration + 1} ===")
    stderr = get_errors()
    errors = parse_errors(stderr)
    
    if not errors:
        print("BUILD PASSES! All errors fixed.")
        break
    
    print(f"Found {len(errors)} files with errors")
    for fname, lineno in sorted(errors):
        print(f"  {fname}:{lineno}")
        fix_file(fname)
else:
    print("Max iterations reached. Some errors may remain.")
    print(stderr)
