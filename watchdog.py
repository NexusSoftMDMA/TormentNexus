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
        "type": "python",
        "critical": True,
    },
    "bobbybookmarks_sync": {
        "script": "scripts/bobbybookmarks_sync.py",
        "args": [],
        "log": "data/bobby_sync_watchdog.log",
        "type": "python",
        "critical": True,
    },
    "trends_analyzer": {
        "script": "scripts/trends_analyzer.py",
        "args": [],
        "log": "data/trends_watchdog.log",
        "type": "python",
        "critical": False,
    },
    # Core TormentNexus services (checked by port)
    "freellm_proxy": {
        "binary": "freellm.exe",
        "port": 4000,
        "type": "port",
        "critical": True,
    },
    "go_sidecar": {
        "binary": "tormentnexus.exe",
        "port": 7778,
        "type": "port",
        "critical": True,
    },
    "go_sidecar_legacy": {
        "binary": "tormentnexus.exe",
        "port": 4300,
        "type": "port",
        "critical": False,
    },
    "ts_control_plane": {
        "binary": "node.exe",
        "port": 4100,
        "type": "port",
        "critical": True,
    },
    "dashboard": {
        "binary": "node.exe",
        "port": 7779,
        "type": "port",
        "critical": True,
    },
    "lm_studio": {
        "binary": "lm-studio.exe",
        "port": 1234,
        "type": "port",
        "critical": False,
    },
    "go_sidecar_old": {
        "binary": "tormentnexus.exe",
        "port": 8080,
        "type": "port",
        "critical": False,
    },
}


def log(msg):
    ts = datetime.now().strftime("%H:%M:%S")
    line = f"[{ts}][WATCHDOG] {msg}"
    with open(LOG_PATH, "a", encoding="utf-8") as f:
        f.write(line + "\n")
    print(line)


def find_process(name, config):
    """Check if a process is running. Supports python scripts and port-checked services."""
    if config.get("type") == "port":
        # Check by port number (works for any service)
        try:
            result = subprocess.run(
                ["netstat", "-ano"],
                capture_output=True,
                text=True,
                timeout=10,
                creationflags=subprocess.CREATE_NO_WINDOW,
            )
            for line in result.stdout.split("\n"):
                if f":{config['port']} " in line and "LISTENING" in line:
                    pid = line.strip().split()[-1]
                    if pid.isdigit():
                        return [int(pid)]
            return []
        except Exception:
            return []
    else:
        # Python script check — use wmic which works without admin rights
        script_name = config["script"]
        try:
            result = subprocess.run(
                [
                    "wmic",
                    "process",
                    "where",
                    'name="pythonw.exe" or name="python.exe"',
                    "get",
                    "ProcessId,CommandLine",
                    "/format:csv",
                ],
                capture_output=True,
                text=True,
                timeout=10,
                creationflags=subprocess.CREATE_NO_WINDOW,
            )
            pids = []
            for line in result.stdout.split("\n"):
                if script_name in line and "python" in line.lower():
                    parts = line.strip().split(",")
                    for part in parts:
                        if part.strip().isdigit():
                            pids.append(int(part.strip()))
                            break
            return pids
        except Exception:
            # Fallback: try tasklist
            try:
                result = subprocess.run(
                    ["tasklist", "/FI", "IMAGENAME eq python.exe", "/FO", "CSV", "/NH"],
                    capture_output=True,
                    text=True,
                    timeout=10,
                    creationflags=subprocess.CREATE_NO_WINDOW,
                )
                lines = [l for l in result.stdout.split("\n") if l.strip()]
                return [
                    int(l.split(",")[1].strip().strip('"')) for l in lines if l.strip()
                ]
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
            creationflags=subprocess.CREATE_NO_WINDOW
            | subprocess.CREATE_NEW_PROCESS_GROUP
            | subprocess.DETACHED_PROCESS,
        )

        log(f"Started {name} (PID {proc.pid})")

        # Write PID file if configured
        pid_file = config.get("pid_file")
        if pid_file:
            pid_path = WORKSPACE / pid_file
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
        pids = find_process(name, config)

        if pids:
            # If more than one, kill duplicates — keep the first one
            if len(pids) > 1:
                for extra in pids[1:]:
                    try:
                        subprocess.run(["taskkill", "/F", "/PID", str(extra)], capture_output=True, timeout=5, creationflags=subprocess.CREATE_NO_WINDOW)
                        log(f"{name}: killed duplicate PID {extra}", "WARN")
                    except Exception:
                        pass
            log(f"{name}: OK (PID {pids[0]})")
        else:
            log(f"{name}: DOWN — restarting...")
            all_ok = False
            if config.get("type") == "port":
                log(
                    f"{name} is a port-based service — cannot auto-restart. Check manually."
                )
            else:
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
        if config.get("type") == "port":
            pids = find_process(name, config)
            if pids:
                log(f"{name}: OK (port {config['port']}, PID {pids[0]})")
            else:
                log(f"{name}: port {config['port']} not listening — check manually")
        else:
            pids = find_process(name, config)
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
