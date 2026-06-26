#!/usr/bin/env python3
"""Remove corruption markers from Go tool files."""
import os
import glob

tool_dir = "go/internal/tools"
patterns = [
    "*deepseek-reasoner*",
    "*openai/gpt*", 
    "*qwen*",
    "*Mistral*",
    "*huggingface*",
    "*nvidia*",
    "*Next-Step Director*",
]

count = 0
for fpath in glob.glob(os.path.join(tool_dir, "*.go")):
    with open(fpath, 'r', encoding='utf-8', errors='replace') as f:
        lines = f.readlines()
    
    modified = False
    new_lines = []
    skip_next = False
    
    for i, line in enumerate(lines):
        trimmed = line.strip()
        
        if trimmed == "---":
            # Check if next line is a model reference
            if i + 1 < len(lines):
                next_trimmed = lines[i + 1].strip()
                is_model_ref = any(p.strip('*') in next_trimmed or 
                                 p.replace('*', '').replace('/', '_') in next_trimmed
                                 for p in patterns)
                if is_model_ref:
                    skip_next = True
                    modified = True
                    continue
            # If we get here, "---" wasn't followed by a model ref, keep it
            new_lines.append(line)
            continue
        
        if skip_next:
            skip_next = False
            modified = True
            continue
        
        new_lines.append(line)
    
    if modified:
        with open(fpath, 'w', encoding='utf-8', newline='') as f:
            f.writelines(new_lines)
        count += 1
        print(f"Fixed: {fpath}")

print(f"\nTotal files fixed: {count}")
