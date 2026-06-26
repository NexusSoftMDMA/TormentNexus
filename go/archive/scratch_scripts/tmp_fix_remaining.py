#!/usr/bin/env python3
"""Fix remaining Go syntax errors in tool files."""
import re

# 1. Fix buildwithclaude.go - misplaced braces in map literal
with open('go/internal/tools/buildwithclaude.go', 'r') as f:
    content = f.read()

# Fix: `return ok(map[string]interface{}{\n}\n\t\t"tools"` -> remove the stray `}`
content = content.replace(
    'return ok(map[string]interface{}{\n}\n\t\t"tools',
    'return ok(map[string]interface{}{\n\t\t"tools'
)
content = content.replace(
    'return ok(map[string]interface{}{\n}\n\t\t"name',
    'return ok(map[string]interface{}{\n\t\t"name'
)

with open('go/internal/tools/buildwithclaude.go', 'w') as f:
    f.write(content)
print('Fixed buildwithclaude.go')

# 2. Fix cws_mcp.go - stray `default:` keyword
with open('go/internal/tools/cws_mcp.go', 'r') as f:
    lines = f.readlines()
# Find stray `default:` or `case:` lines that are outside any switch
# These are typically at the top level (outside a function body)
new_lines = []
for i, line in enumerate(lines):
    stripped = line.strip()
    # If we see a stray case/default at file level (not inside a function), remove it
    # Heuristic: if previous non-blank line was a function/switch, keep it; 
    # if it looks orphaned, remove it
    new_lines.append(line)

with open('go/internal/tools/cws_mcp.go', 'r') as f:
    content = f.read()
print('cws_mcp.go content:')
for i, line in enumerate(content.split('\n')[:30]):
    print(f'{i+1}: {line}')
print()

# 3. Fix decoder_3am_mcp.go - stray `case:` keyword
with open('go/internal/tools/decoder_3am_mcp.go', 'r') as f:
    content = f.read()
print('decoder_3am_mcp.go content:')
for i, line in enumerate(content.split('\n')[:30]):
    print(f'{i+1}: {line}')
print()

# 4. Fix dreb_semantic_search.go - newline in argument list
with open('go/internal/tools/dreb_semantic_search.go', 'r') as f:
    content = f.read()
# Look for incomplete function call around line 48
lines = content.split('\n')
for i in range(45, min(55, len(lines))):
    print(f'{i+1}: {lines[i]}')
print()

# 5. Fix dversum_mcp_server.go - unexpected }
with open('go/internal/tools/dversum_mcp_server.go', 'r') as f:
    content = f.read()
print('dversum_mcp_server.go content:')
for i, line in enumerate(content.split('\n')[:35]):
    print(f'{i+1}: {line}')
