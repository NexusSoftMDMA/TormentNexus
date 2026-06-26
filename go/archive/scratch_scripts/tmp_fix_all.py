#!/usr/bin/env python3
"""Comprehensive Go syntax fixer for swarm-generated files."""
import re, os, glob, subprocess

tool_dir = "go/internal/tools"

files_fixed = set()

# === Pattern fixes ===

# 1. buildwithclaude.go: unclosed function before HandleGetTool
with open(f'{tool_dir}/buildwithclaude.go', 'r') as f:
    content = f.read()
# The function HandleListTools ends prematurely - the `}` after the map literal 
# closes the function early. Fix: add proper closing to HandleListTools
old = '''	return ok(map[string]interface{}{
		"tools": []string{"tool1", "tool2"},
	})
'''
if old in content:
    # The closing `})` is correct, but there might be missing function close
    content = content.replace(
        '})', 
        '})\n}', 1
    )
    # Also remove the extra doubled line
    with open(f'{tool_dir}/buildwithclaude.go', 'w') as f:
        f.write(content)
    files_fixed.add('buildwithclaude.go')
    print('Fixed buildwithclaude.go')

# 2. decoder_3am_mcp.go: unexpected EOF
with open(f'{tool_dir}/decoder_3am_mcp.go', 'r') as f:
    lines = f.readlines()
# Check if last non-blank line is `}` - this should close both switch and function,
# but the switch has 3 cases, so there should be a switch close `}` AND a func close `}`
# If file ends with single `}`, add another one
end_lines = [l.rstrip() for l in lines[-5:] if l.strip()]
if end_lines and end_lines[-1] == '}':
    # Count opening braces vs closing braces
    opens = content.count('{')
    closes = content.count('}')
    if opens > closes:
        # Need more closing braces
        content = content.rstrip() + '\n}\n'
        with open(f'{tool_dir}/decoder_3am_mcp.go', 'w') as f:
            f.write(content)
        files_fixed.add('decoder_3am_mcp.go')
        print('Fixed decoder_3am_mcp.go (added closing brace)')
    
# 3. cws_mcp.go: also check similar issue
with open(f'{tool_dir}/cws_mcp.go', 'r') as f:
    content2 = f.read()
opens = content2.count('{')
closes = content2.count('}')
if opens > closes:
    content2 = content2.rstrip() + '\n}\n'
    with open(f'{tool_dir}/cws_mcp.go', 'w') as f:
        f.write(content2)
    files_fixed.add('cws_mcp.go')
    print('Fixed cws_mcp.go (added closing brace)')

# === Fix common patterns across ALL files ===

# Find incomplete return ok(err(... statements (where the argument is incomplete)
for fpath in glob.glob(f'{tool_dir}/*.go'):
    with open(fpath, 'r', encoding='utf-8', errors='replace') as f:
        content = f.read()
    
    original = content
    changed = False
    
    # Pattern: `return ok("..." +` at end of a line (incomplete concatenation)
    content = re.sub(
        r'return ok\("[^"]*" \+$', 
        lambda m: m.group() + ' string(body))', 
        content, 
        flags=re.MULTILINE
    )
    # Catch the case without closing paren: `return ok("text " +` 
    content = re.sub(
        r'(return ok\("[^"]*" \+)\s*$', 
        r'\1 string(body))', 
        content, 
        flags=re.MULTILINE
    )
    
    # Pattern: `return ok(string(body))` line that has `return ok(body)` - fix
    content = re.sub(
        r'return ok\(body\)', 
        'return ok(string(body))', 
        content
    )
    
    # Pattern: `return ok` without argument (should be `return ok(string(body))`)
    content = re.sub(
        r'return ok(?![(])',  # return ok NOT followed by (
        'return ok(string(body))', 
        content
    )
    
    if content != original:
        with open(fpath, 'w', encoding='utf-8', newline='') as f:
            f.write(content)
        files_fixed.add(os.path.basename(fpath))

print(f'\nTotal files fixed: {len(files_fixed)}')
for f in sorted(files_fixed):
    print(f'  {f}')
