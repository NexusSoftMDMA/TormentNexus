"""
TormentNexus Watchdog
=====================
Ensures all workers are always running. If any process dies, restarts it.
Checks every 60 seconds. Runs in the background.

Monitored workers:
  - swarm_v7.py (5 workers, freellm-only code generation)
  - bobbybookmarks_sync.py (hourly bookmark sync)
  - trends_analyzer.py (6-hour trend analysis)
"""

import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

WORKSPACE = Path(__file__).resolve().parent
LOG_PATH = WORKSPACE / "data" / "watchdog.log"
CHECK_INTERVAL = 60  # seconds between health checks

WORKERS = {
    "swarm": {
        "script": "swarm_v7.py",
        "args": ["--forever"],
        "log": "data/swarm_watchdog.log",
        "pid_file": "swarm_forever.pid",
        "critical": True,
    },
    "bobbybookmarks_sync": {
        "script": "scripts/bobbybookmarks_sync.py",
        "args": [],
        "log": "data/bobby_sync_watchdog.log",
        "pid_file": None,
        "critical": True,
    },
    "trends_analyzer": {
        "script": "scripts/trends_analyzer.py",
        "args": [],
        "log": "data/trends_watchdog.log",
        "pid_file": None,
        "critical": False,
    },
}

def log(msg):
    ts = datetime.now().strftime("%H:%M:%S")
    line = f"[{ts}][WATCHDOG] {msg}"
    with open(LOG_PATH, "a", encoding="utf-8") as f:
        f.write(line + "\n")
    print(line)

def find_process(script_name):
    """Check if a Python process running the given script exists."""
    try:
        result = subprocess.run(
            ['powershell', '-Command',
             f'Get-Process python* | Where-Object {{ $_.CommandLine -match "{script_name}" }} | Select-Object -ExpandProperty Id'],
            capture_output=True, text=True, timeout=10,
        )
        pids = result.stdout.strip().split()
        return [int(p) for p in pids if p.isdigit()]
    except Exception:
        return []

def start_worker(name, config):
    """Start a worker process. Returns the PID or None."""
    script_path = WORKSPACE / config["script"]
    log_path = str(WORKSPACE / config["log"])
    
    if not script_path.exists():
        log(f"ERROR: {script_path} not found — cannot start {name}")
        return None
    
    try:
        cmd = [sys.executable, "-u", str(script_path)] + config["args"]
        logfile = open(log_path, "a", buffering=1)
        
        proc = subprocess.Popen(
            cmd,
            stdout=logfile,
            stderr=subprocess.STDOUT,
            cwd=str(WORKSPACE),
            creationflags=subprocess.CREATE_NEW_PROCESS_GROUP | subprocess.DETACHED_PROCESS,
        )
        
        log(f"Started {name} (PID {proc.pid})")
        
        # Write PID file if configured
        if config["pid_file"]:
            pid_path = WORKSPACE / config["pid_file"]
            with open(pid_path, "w") as f:
                f.write(str(proc.pid))
        
        return proc.pid
    except Exception as e:
        log(f"Failed to start {name}: {e}")
        return None

def check_and_repair():
    """Check all workers and restart any that died."""
    all_ok = True
    
    for name, config in WORKERS.items():
        pids = find_process(config["script"])
        
        if pids:
            log(f"{name}: OK (PID{'/'.join(str(p) for p in pids)})")
        else:
            log(f"{name}: DOWN — restarting...")
            all_ok = False
            pid = start_worker(name, config)
            if pid:
                log(f"{name}: restarted successfully (PID {pid})")
            else:
                log(f"{name}: FAILED to restart")
    
    return all_ok

def main():
    log("=" * 60)
    log("TORMENTNEXUS WATCHDOG STARTED")
    log(f"Monitoring {len(WORKERS)} workers")
    log(f"Check interval: {CHECK_INTERVAL}s")
    log("=" * 60)
    
    # Initial startup: start all workers
    for name, config in WORKERS.items():
        pids = find_process(config["script"])
        if not pids:
            log(f"Starting {name} for first time...")
            start_worker(name, config)
        else:
            log(f"{name} already running (PID{'/'.join(str(p) for p in pids)})")
    
    cycles = 0
    while True:
        time.sleep(CHECK_INTERVAL)
        cycles += 1
        
        log(f"--- Health check #{cycles} ---")
        all_ok = check_and_repair()
        
        if all_ok:
            log("All workers healthy")
        else:
            log("Some workers were repaired")
        
        # Rotate log every 1000 checks (~16.6 hours)
        if cycles % 1000 == 0:
            log(f"Watchdog running for {cycles * CHECK_INTERVAL / 3600:.1f} hours")
            log(f"Last check: {datetime.now().isoformat()}")

if __name__ == "__main__":
    main()
