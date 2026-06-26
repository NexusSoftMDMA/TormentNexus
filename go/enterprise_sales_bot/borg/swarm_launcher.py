"""Persistent swarm launcher - keeps swarm alive even if parent shell exits."""

import subprocess
import sys
import os

log_path = os.path.join(os.getcwd(), "data", "swarm_v7_bg.log")
pid_path = os.path.join(os.getcwd(), "swarm_forever.pid")

log = open(log_path, "w", buffering=1)

proc = subprocess.Popen(
    [sys.executable, "-u", "swarm_v7.py", "--workers", "3", "--forever"],
    stdout=log,
    stderr=subprocess.STDOUT,
    cwd=os.getcwd(),
    creationflags=subprocess.CREATE_NEW_PROCESS_GROUP | subprocess.DETACHED_PROCESS,
)

with open(pid_path, "w") as f:
    f.write(str(proc.pid))

print(f"Swarm launched: PID {proc.pid}", flush=True)
print(f"Log: {log_path}", flush=True)
