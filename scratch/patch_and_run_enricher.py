import os
import sys
import subprocess

brain_scratch_dir = r"C:\Users\hyper\.gemini\antigravity\brain\e88bac4f-e064-4c4b-bf5f-17f3373dac43\scratch"
workspace_scratch_dir = r"c:\Users\hyper\workspace\borg\scratch"
target_db = r"c:\Users\hyper\workspace\borg\tormentnexus.db"

script_name = "enrich_metadata.py"
src_path = os.path.join(brain_scratch_dir, script_name)

if not os.path.exists(src_path):
    print(f"Error: {src_path} not found!")
    sys.exit(1)

print(f"=== PATCHING {script_name} ===")
with open(src_path, "r", encoding="utf-8") as f:
    content = f.read()

# Replace db path
patched = content.replace("borg.db", "tormentnexus.db")
patched = patched.replace('BORG_DB_PATH = r"c:\\Users\\hyper\\workspace\\borg\\borg.db"', f'BORG_DB_PATH = r"{target_db}"')

# Adjust GitHub API enrichment limit to prevent long rate-limiting delays
patched = patched.replace("LIMIT 5000", "LIMIT 50")

# Save to workspace scratch dir
dest_path = os.path.join(workspace_scratch_dir, "patched_" + script_name)
with open(dest_path, "w", encoding="utf-8") as f:
    f.write(patched)
print(f"Patched script saved to {dest_path}")

print("=== RUNNING ENRICHER ===")
try:
    res = subprocess.run([sys.executable, dest_path], capture_output=True, text=True)
    print("STDOUT:")
    print(res.stdout)
    if res.stderr:
        print("STDERR:")
        print(res.stderr)
except Exception as e:
    print(f"Execution failed: {e}")

print("=== PATCH AND RUN COMPLETE ===")
