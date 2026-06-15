#!/usr/bin/env python3
"""
TormentNexus Swarm v4 — THE INFINITE ASSIMILATOR
=================================================

A resilient multi-agent system that continuously assimilates MCP servers
into native Go modules using the FreeLLM proxy.

DESIGN PRINCIPLES:
1. SERIAL per-worker — one LLM call at a time per worker
2. AGGRESSIVE retry — exponential backoff with circuit breaker
3. MANIFEST-based registry merge — no concurrent registry.go writes
4. BUILD verification — auto-remove broken files, fix registry refs
5. PROXY auto-restart — kills and restarts when death-spiraling
6. PERSISTENT state — SQLite tracks task attempts, survive crashes
7. NEVER delete core files — registry.go, server.go, parity.go protected

KEY INSIGHT: The FreeLLM proxy handles ~2 concurrent requests.
Each LLM call takes 85-150s. So 2 workers = max throughput.

USAGE:
  python3 swarm_v4.py                        # Default: 2 workers, 50 tasks
  python3 swarm_v4.py --workers 4 --limit 200
  python3 swarm_v4.py --forever              # Run until all pending done
  python3 swarm_v4.py --repair               # Fix broken files only
"""

import argparse
import json
import os
import re
import signal
import sqlite3
import subprocess
import sys
import textwrap
import threading
import time
import traceback
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime
from pathlib import Path
from typing import Optional

# ══════════════════════════════════════════════════════════════════════════════
# CONFIGURATION
# ══════════════════════════════════════════════════════════════════════════════

PROXY_URL = os.environ.get("SWARM_PROXY", "http://localhost:4000")
PROXY_KEY = os.environ.get("SWARM_KEY", "sk-freellm")
PROXY_BIN = os.environ.get(
    "SWARM_PROXY_BIN", "C:/Users/hyper/workspace/litellm_control_panel/freellm.exe"
)
WORKSPACE = Path(
    os.environ.get("SWARM_WORKSPACE", "C:/Users/hyper/workspace/tormentnexus")
)
GO_DIR = WORKSPACE / "go"
TOOLS_DIR = GO_DIR / "internal" / "tools"
REGISTRY_FILE = TOOLS_DIR / "registry.go"
DB_PATH = WORKSPACE / "data" / "assimilation_state.db"
MANIFEST_DIR = TOOLS_DIR / "manifests"

# Protected files that should NEVER be deleted by build repair
PROTECTED_FILES = {
    "registry.go",
    "parity.go",
    "server.go",
    "factory.go",
    "basic_memory.go",
    "filesystem.go",
    "web_fetch.go",
    "sqlite.go",
    "bash.go",
    "glob.go",
    "apply_patch.go",
    "multi_edit.go",
    "git_ingest.go",
}

# Models tiered by reliability and speed — only use free/fast ones
MODEL_TIERS = [
    # Tier 1: Fast and reliable (flash models)
    ["gemini-3.5-flash", "google/gemini-2.5-flash", "google/gemini-2.5-flash-lite"],
    # Tier 2: Good quality, slower
    [
        "deepseek/deepseek-v4-flash",
        "deepseek-v4-flash-free",
        "deepseek-ai/deepseek-v4-flash",
    ],
    # Tier 3: Decent alternatives
    ["qwen/qwen3.6-flash", "stepfun/step-3.5-flash", "xiaomi/mimo-v2-flash"],
    # Tier 4: Fallback
    ["gpt-4o-mini", "openai/gpt-4o-mini", "free-llm"],
]

# Flatten for rotation
ALL_MODELS = [m for tier in MODEL_TIERS for m in tier]

# Retry configuration
MAX_RETRIES = 5
RETRY_BASE_DELAY = 60  # seconds
RETRY_MAX_DELAY = 600  # max 10 minutes
CIRCUIT_BREAKER_THRESHOLD = 3  # consecutive failures
CIRCUIT_BREAKER_RESET = 300  # 5 minutes cooldown
PROXY_RESTART_THRESHOLD = 5  # consecutive proxy errors
BUILD_BATCH_SIZE = 5  # verify build every N tasks

# ══════════════════════════════════════════════════════════════════════════════
# LOGGING
# ══════════════════════════════════════════════════════════════════════════════


class Logger:
    """Thread-safe colored logger."""

    COLORS = {
        "RED": "\033[91m",
        "GREEN": "\033[92m",
        "YELLOW": "\033[93m",
        "BLUE": "\033[94m",
        "MAGENTA": "\033[95m",
        "CYAN": "\033[96m",
        "RESET": "\033[0m",
    }

    def __init__(self, log_file=None):
        self.lock = threading.Lock()
        self.log_file = log_file
        self.start_time = time.time()
        self.stats = {
            "tasks_completed": 0,
            "tasks_failed": 0,
            "builds_clean": 0,
            "builds_broken": 0,
            "proxy_restarts": 0,
            "llm_calls": 0,
            "llm_errors": 0,
        }

    def _ts(self):
        elapsed = int(time.time() - self.start_time)
        h, m, s = elapsed // 3600, (elapsed % 3600) // 60, elapsed % 60
        return f"{h:02d}:{m:02d}:{s:02d}"

    def _log(self, level, color, worker, msg):
        ts = self._ts()
        prefix = f"[{ts}]"
        if worker is not None:
            prefix += f" [W{worker}]"
        prefix += f" [{level}]"
        line = f"{prefix} {msg}"

        with self.lock:
            c = self.COLORS.get(color, "")
            print(f"{c}{line}{self.COLORS['RESET']}", flush=True)
            if self.log_file:
                try:
                    with open(
                        self.log_file, "a", encoding="utf-8", errors="replace"
                    ) as f:
                        f.write(line + "\n")
                except Exception:
                    pass

    def info(self, msg, worker=None):
        self._log("INFO", "CYAN", worker, msg)

    def success(self, msg, worker=None):
        self._log("OK", "GREEN", worker, msg)

    def warn(self, msg, worker=None):
        self._log("WARN", "YELLOW", worker, msg)

    def error(self, msg, worker=None):
        self._log("ERR", "RED", worker, msg)

    def llm(self, msg, worker=None):
        self._log("LLM", "MAGENTA", worker, msg)
        with self.lock:
            self.stats["llm_calls"] += 1

    def llm_err(self, msg, worker=None):
        self._log("LLM-ERR", "RED", worker, msg)
        with self.lock:
            self.stats["llm_errors"] += 1

    def status(self, msg):
        self._log("STATUS", "BLUE", None, msg)

    def inc_completed(self):
        with self.lock:
            self.stats["tasks_completed"] += 1

    def inc_failed(self):
        with self.lock:
            self.stats["tasks_failed"] += 1

    def stats_summary(self):
        with self.lock:
            s = self.stats.copy()
        elapsed = int(time.time() - self.start_time)
        return (
            f"Completed: {s['tasks_completed']} | Failed: {s['tasks_failed']} | "
            f"LLM calls: {s['llm_calls']} | LLM errors: {s['llm_errors']} | "
            f"Proxy restarts: {s['proxy_restarts']} | "
            f"Elapsed: {elapsed // 3600}h{(elapsed % 3600) // 60}m"
        )


log = Logger()

# ══════════════════════════════════════════════════════════════════════════════
# PROXY MANAGEMENT
# ══════════════════════════════════════════════════════════════════════════════


class ProxyManager:
    """Manages the FreeLLM proxy lifecycle with health checks and auto-restart."""

    def __init__(self):
        self.consecutive_errors = 0
        self.last_restart = 0
        self.lock = threading.Lock()
        self._is_healthy = True

    def check_health(self):
        """Check proxy health endpoint."""
        try:
            import urllib.request

            req = urllib.request.Request(f"{PROXY_URL}/health", method="GET")
            with urllib.request.urlopen(req, timeout=10) as resp:
                if resp.status == 200:
                    self._is_healthy = True
                    return True
        except Exception:
            pass
        self._is_healthy = False
        return False

    def restart(self):
        """Kill and restart the proxy."""
        with self.lock:
            if time.time() - self.last_restart < 30:
                log.warn("Proxy restart too recent, waiting 30s...")
                time.sleep(30)

            log.status("Killing proxy...")
            try:
                subprocess.run(
                    ["taskkill", "//F", "//IM", "freellm.exe"],
                    capture_output=True,
                    timeout=10,
                )
            except Exception:
                try:
                    subprocess.run(
                        ["pkill", "-f", "freellm"],
                        capture_output=True,
                        timeout=5,
                    )
                except Exception:
                    pass

            time.sleep(15)

            log.status("Starting proxy...")
            try:
                proxy_dir = str(Path(PROXY_BIN).parent)
                subprocess.Popen(
                    [PROXY_BIN],
                    cwd=proxy_dir,
                    stdout=subprocess.DEVNULL,
                    stderr=subprocess.DEVNULL,
                )
            except Exception as e:
                log.error(f"Failed to start proxy: {e}")
                return False

            for attempt in range(30):
                time.sleep(5)
                if self.check_health():
                    log.success("Proxy restarted and healthy!")
                    self.consecutive_errors = 0
                    self.last_restart = time.time()
                    with log.lock:
                        log.stats["proxy_restarts"] += 1
                    return True
                log.warn(f"Proxy not healthy yet (attempt {attempt + 1}/30)...")

            log.error("Proxy failed to come up after restart")
            return False

    def record_error(self):
        """Record a proxy-related error."""
        with self.lock:
            self.consecutive_errors += 1
            if self.consecutive_errors >= PROXY_RESTART_THRESHOLD:
                log.error(f"Proxy error threshold ({PROXY_RESTART_THRESHOLD}) reached!")
                self.restart()

    def record_success(self):
        """Record a successful proxy call."""
        with self.lock:
            self.consecutive_errors = 0

    def is_available(self):
        return self._is_healthy


proxy_mgr = ProxyManager()

# ══════════════════════════════════════════════════════════════════════════════
# LLM CLIENT
# ══════════════════════════════════════════════════════════════════════════════


class LLMClient:
    """Thread-safe LLM client with retry, circuit breaker, and model rotation."""

    def __init__(self):
        self.model_index = 0
        self.lock = threading.Lock()
        self.circuit_open = False
        self.circuit_opened_at = 0
        self.consecutive_failures = 0

    def _get_next_model(self):
        with self.lock:
            model = ALL_MODELS[self.model_index % len(ALL_MODELS)]
            self.model_index += 1
            return model

    def _check_circuit(self):
        if self.circuit_open:
            if time.time() - self.circuit_opened_at > CIRCUIT_BREAKER_RESET:
                log.status("Circuit breaker reset — trying again...")
                self.circuit_open = False
                self.consecutive_failures = 0
                return True
            return False
        return True

    def _trip_circuit(self):
        self.circuit_open = True
        self.circuit_opened_at = time.time()
        log.warn(f"Circuit breaker OPEN — cooling down {CIRCUIT_BREAKER_RESET}s")

    def call(
        self,
        prompt: str,
        system: str = "",
        model: Optional[str] = None,
        worker_id: int = 0,
        max_retries: int = MAX_RETRIES,
    ) -> Optional[str]:
        """Call the LLM with aggressive retry. Returns response text or None."""
        if not self._check_circuit():
            return None

        chosen_model = model or self._get_next_model()

        for attempt in range(max_retries):
            try:
                import urllib.request

                payload = json.dumps(
                    {
                        "model": chosen_model,
                        "messages": [
                            {"role": "system", "content": system},
                            {"role": "user", "content": prompt},
                        ],
                        "temperature": 0.3,
                        "max_tokens": 8192,
                    }
                ).encode("utf-8")

                req = urllib.request.Request(
                    f"{PROXY_URL}/v1/chat/completions",
                    data=payload,
                    headers={
                        "Content-Type": "application/json",
                        "Authorization": f"Bearer {PROXY_KEY}",
                    },
                    method="POST",
                )

                timeout = 180 + (attempt * 60)
                log.llm(
                    f"Calling {chosen_model} (attempt {attempt + 1}/{max_retries}, "
                    f"timeout={timeout}s)",
                    worker_id,
                )

                with urllib.request.urlopen(req, timeout=timeout) as resp:
                    data = json.loads(resp.read().decode("utf-8"))
                    content = data["choices"][0]["message"]["content"]
                    self.consecutive_failures = 0
                    proxy_mgr.record_success()
                    return content

            except Exception as e:
                err_str = str(e)[:200]
                log.llm_err(f"{chosen_model}: {err_str}", worker_id)

                self.consecutive_failures += 1
                proxy_mgr.record_error()

                if self.consecutive_failures >= CIRCUIT_BREAKER_THRESHOLD:
                    self._trip_circuit()
                    return None

                base_delay = min(RETRY_BASE_DELAY * (2**attempt), RETRY_MAX_DELAY)
                jitter = attempt * 10
                delay = base_delay + jitter

                is_proxy_err = any(
                    x in err_str.lower()
                    for x in [
                        "502",
                        "503",
                        "connectionreset",
                        "connection refused",
                        "timed out",
                        "no models",
                        "reset",
                        "read error",
                    ]
                )
                if is_proxy_err:
                    log.warn(f"Proxy error — waiting {delay}s", worker_id)
                    chosen_model = self._get_next_model()
                else:
                    delay = min(delay, 30)

                time.sleep(delay)

        log.error(f"All {max_retries} retries exhausted for {chosen_model}", worker_id)
        self._trip_circuit()
        return None


llm = LLMClient()

# ══════════════════════════════════════════════════════════════════════════════
# DATABASE OPERATIONS — Uses the MAIN assimilation DB
# ══════════════════════════════════════════════════════════════════════════════


def get_pending_tasks(limit=50):
    """Get pending MCP servers from the state DB, prioritized by score."""
    db = sqlite3.connect(str(DB_PATH))
    c = db.cursor()
    rows = c.execute(
        """
        SELECT name, github_url, score, category, description, subcategory, tags
        FROM mcp_servers 
        WHERE status='pending'
        ORDER BY score DESC
        LIMIT ?
    """,
        (limit,),
    ).fetchall()
    db.close()

    tasks = []
    for row in rows:
        tasks.append(
            {
                "name": row[0],
                "github_url": row[1] or "",
                "score": row[2] or 0,
                "category": row[3] or "unknown",
                "description": row[4] or "",
                "subcategory": row[5] or "",
                "tags": row[6] or "",
            }
        )
    return tasks


def get_retry_tasks(limit=20):
    """Get previously failed tasks worth retrying (from main DB)."""
    db = sqlite3.connect(str(DB_PATH))
    c = db.cursor()
    rows = c.execute(
        """
        SELECT name, github_url, score, category, description, subcategory, tags
        FROM mcp_servers 
        WHERE status='failed'
        ORDER BY score DESC
        LIMIT ?
    """,
        (limit,),
    ).fetchall()
    db.close()

    tasks = []
    for row in rows:
        tasks.append(
            {
                "name": row[0],
                "github_url": row[1] or "",
                "score": row[2] or 0,
                "category": row[3] or "unknown",
                "description": row[4] or "",
                "subcategory": row[5] or "",
                "tags": row[6] or "",
                "is_retry": True,
            }
        )
    return tasks


def mark_server_implemented(server_name, go_file, tools_count=0):
    """Mark a server as implemented in the main DB."""
    db = sqlite3.connect(str(DB_PATH))
    c = db.cursor()
    c.execute(
        """
        UPDATE mcp_servers SET status='implemented', go_file=?, tools_exposed=?,
        updated_at=CURRENT_TIMESTAMP WHERE name=?
    """,
        (go_file, str(tools_count), server_name),
    )
    db.commit()
    db.close()


def mark_server_failed(server_name, error=""):
    """Mark a server as failed in the main DB."""
    db = sqlite3.connect(str(DB_PATH))
    c = db.cursor()
    c.execute(
        """
        UPDATE mcp_servers SET status='failed', notes=?,
        updated_at=CURRENT_TIMESTAMP WHERE name=?
    """,
        (error[:500], server_name),
    )
    db.commit()
    db.close()


def mark_server_pending(server_name):
    """Reset a failed server back to pending for retry."""
    db = sqlite3.connect(str(DB_PATH))
    c = db.cursor()
    c.execute(
        """
        UPDATE mcp_servers SET status='pending', notes='retry',
        updated_at=CURRENT_TIMESTAMP WHERE name=?
    """,
        (server_name,),
    )
    db.commit()
    db.close()


def get_db_stats():
    """Get current DB statistics."""
    db = sqlite3.connect(str(DB_PATH))
    c = db.cursor()
    total = c.execute("SELECT COUNT(*) FROM mcp_servers").fetchone()[0]
    impl = c.execute(
        'SELECT COUNT(*) FROM mcp_servers WHERE status="implemented"'
    ).fetchone()[0]
    pend = c.execute(
        'SELECT COUNT(*) FROM mcp_servers WHERE status="pending"'
    ).fetchone()[0]
    fail = c.execute(
        'SELECT COUNT(*) FROM mcp_servers WHERE status="failed"'
    ).fetchone()[0]
    db.close()
    return total, impl, pend, fail


# ══════════════════════════════════════════════════════════════════════════════
# GO CODE GENERATION
# ══════════════════════════════════════════════════════════════════════════════


def server_name_to_go_filename(server_name):
    """Convert server name to a valid Go filename."""
    name = server_name
    for prefix in ["mcp-server-", "mcp-", "@", "server-", "mcp_"]:
        if name.lower().startswith(prefix):
            name = name[len(prefix) :]
    name = re.sub(r"[/.\-]+", "_", name)
    name = re.sub(r"[^a-zA-Z0-9_]", "", name)
    name = name.lower()
    if not name or len(name) < 2:
        name = f"tool_{hash(server_name) % 10000}"
    # Max filename length
    if len(name) > 60:
        name = name[:60]
    return name


def build_prompt(task):
    """Build the LLM prompt for generating a Go tool implementation."""
    name = task["name"]
    url = task.get("github_url", "")
    desc = task.get("description", "")
    cat = task.get("category", "")
    tags = task.get("tags", "")
    subcat = task.get("subcategory", "")
    go_filename = server_name_to_go_filename(name)

    prompt = textwrap.dedent(f"""\
    You are a Go engineer implementing native MCP tool handlers for the TormentNexus control plane.
    
    TASK: Implement a Go-native tool module for the MCP server "{name}".
    
    Server info:
    - Name: {name}
    - GitHub: {url}
    - Category: {cat} / {subcat}
    - Description: {desc}
    - Tags: {tags}
    
    REQUIREMENTS:
    1. Package must be `package tools`
    2. Each handler function signature: `func HandleXxx(ctx context.Context, args map[string]interface{{}}) (ToolResponse, error)`
    3. Use the helper functions: `ok(text string) (ToolResponse, error)` for success, `err(text string) (ToolResponse, error)` for error responses
    4. Use `getString(args, "key")`, `getInt(args, "key")`, `getBool(args, "key")` to extract arguments. These return single values.
    5. Use `http.Client{{Timeout: 30 * time.Second}}` for HTTP calls
    6. Use `context.Context` for cancellation
    7. Return JSON strings for structured data
    8. Every function MUST compile — no undefined references, no pseudocode, no TODOs
    9. Do NOT import packages you don't use
    10. Keep it simple — 2-6 handler functions per file
    
    CRITICAL RULES — READ CAREFULLY:
    - `err()` and `ok()` are RESPONSE BUILDERS that return (ToolResponse, error). 
      Example: `return ok("success message")` or `return err("error message")`
    - For error CHECKING use: `if e != nil {{ return err(e.Error()) }}`
    - NEVER call `err()` like `err(response)` — `err` takes a string description, not a response
    - `getString` returns a string (single value), NOT a tuple. Example: `val := getString(args, "key")`
    - `getInt` returns an int (single value). Example: `val := getInt(args, "key")`
    - `getBool` returns a bool (single value). Example: `val := getBool(args, "key")`
    - Do NOT reference any external MCP SDK packages
    - ALL imports must be from the standard library ONLY:
      "context", "encoding/json", "fmt", "io", "net/http", "net/url", "os",
      "os/exec", "path/filepath", "strconv", "strings", "time", "regexp", "sort"
    
    OUTPUT FORMAT — you MUST output TWO sections:
    
    ===GO_FILE===
    (The complete Go source code, starting with `package tools`)
    
    ===MANIFEST===
    (A JSON object with this structure:)
    {{
      "filename": "{go_filename}.go",
      "server_name": "{name}",
      "handlers": [
        {{"tool_name": "tool_name_1", "handler_func": "HandleToolName1", "description": "What it does"}},
        {{"tool_name": "tool_name_2", "handler_func": "HandleToolName2", "description": "What it does"}}
      ]
    }}
    """)

    system = textwrap.dedent("""\
    You are an expert Go engineer who writes clean, compilable code.
    You implement MCP tool handlers for the TormentNexus control plane.
    Every line you write MUST compile without errors.
    You NEVER use pseudocode, TODOs, or placeholder implementations.
    You ALWAYS handle errors properly with `if e != nil { return err(e.Error()) }`.
    The functions `ok(text)` and `err(text)` return (ToolResponse, error).
    getString/getInt/getBool return SINGLE values, not tuples.
    """)

    return prompt, system


def parse_llm_output(output: str, go_filename: str):
    """Parse the LLM output into Go code and manifest."""
    if not output:
        return None, None

    go_code = None
    manifest = None

    # Try ===GO_FILE=== / ===MANIFEST=== sections
    go_match = re.search(
        r"===GO_FILE===\s*\n(.*?)(?====MANIFEST===|$)", output, re.DOTALL
    )
    manifest_match = re.search(r"===MANIFEST===\s*\n(.*?)$", output, re.DOTALL)

    if go_match:
        go_code = go_match.group(1).strip()
        go_code = re.sub(r"^```go\s*\n?", "", go_code)
        go_code = re.sub(r"\n?```\s*$", "", go_code)

    if manifest_match:
        manifest_str = manifest_match.group(1).strip()
        manifest_str = re.sub(r"^```json\s*\n?", "", manifest_str)
        manifest_str = re.sub(r"\n?```\s*$", "", manifest_str)
        try:
            manifest = json.loads(manifest_str)
        except json.JSONDecodeError:
            manifest_str = re.sub(r",\s*}", "}", manifest_str)
            manifest_str = re.sub(r",\s*]", "]", manifest_str)
            try:
                manifest = json.loads(manifest_str)
            except json.JSONDecodeError:
                log.warn("Failed to parse manifest JSON")

    # Fallbacks
    if not go_code:
        code_match = re.search(r"```go\s*\n(.*?)```", output, re.DOTALL)
        if code_match:
            go_code = code_match.group(1).strip()

    if not go_code:
        pkg_match = re.search(r"(package tools\s+.*)", output, re.DOTALL)
        if pkg_match:
            go_code = pkg_match.group(1).strip()

    # Validate
    if go_code and not go_code.strip().startswith("package tools"):
        idx = go_code.find("package tools")
        if idx >= 0:
            go_code = go_code[idx:]
        else:
            go_code = None

    # Build manifest from handler names if missing
    if go_code and not manifest:
        handlers = re.findall(r"func (Handle\w+)\s*\(", go_code)
        manifest = {
            "filename": f"{go_filename}.go",
            "handlers": [
                {
                    "tool_name": re.sub(r"([A-Z])", r"_\1", h[6:]).lower().lstrip("_"),
                    "handler_func": h,
                    "description": h,
                }
                for h in handlers
            ],
        }

    return go_code, manifest


def validate_go_code(code: str) -> list:
    """Basic validation of generated Go code."""
    issues = []
    if not code:
        issues.append("Empty code")
        return issues
    if not code.strip().startswith("package tools"):
        issues.append("Missing 'package tools'")

    bad = [
        (r"(?<!return )err\(\)", "Bare err() call — should be 'return err(msg)'"),
        (r'import\s+"github\.com/(?!tormentnexushq)', "External package import"),
        (r"TODO|FIXME|HACK", "Placeholder comment"),
    ]
    for pattern, desc in bad:
        if re.search(pattern, code):
            issues.append(f"Suspicious: {desc}")
    return issues


def write_go_file(code: str, filename: str) -> bool:
    """Write Go code to a file."""
    filepath = TOOLS_DIR / filename
    try:
        with open(filepath, "w", encoding="utf-8") as f:
            f.write(code)
        return True
    except Exception as e:
        log.error(f"Failed to write {filename}: {e}")
        return False


def write_manifest(manifest: dict, filename: str) -> bool:
    """Write manifest to a JSON file."""
    MANIFEST_DIR.mkdir(exist_ok=True)
    manifest_path = MANIFEST_DIR / f"{filename}.json"
    try:
        with open(manifest_path, "w", encoding="utf-8") as f:
            json.dump(manifest, f, indent=2)
        return True
    except Exception as e:
        log.error(f"Failed to write manifest: {e}")
        return False


# ══════════════════════════════════════════════════════════════════════════════
# BUILD VERIFICATION & AUTO-FIX — PROTECTS CORE FILES
# ══════════════════════════════════════════════════════════════════════════════


def verify_build() -> tuple:
    """
    Run go build. Return (success, broken_files, error_count).
    Only removes tool files (NOT registry.go or protected files).
    """
    try:
        result = subprocess.run(
            ["go", "build", "-buildvcs=false", "./cmd/tormentnexus"],
            capture_output=True,
            text=True,
            cwd=str(GO_DIR),
            timeout=120,
            encoding="utf-8",
            errors="replace",
        )

        if result.returncode == 0:
            log.success("Build CLEAN")
            with log.lock:
                log.stats["builds_clean"] += 1
            return True, [], 0

        error_lines = (result.stderr or "").strip().split("\n")
        broken_tool_files = set()
        undefined_handlers = set()

        for line in error_lines:
            # Broken tool files ONLY (not registry.go or httpapi/)
            match = re.match(r"(internal[\\/]tools[\\/])(\w[\w_]*\.go)", line)
            if match:
                fname = match.group(2)
                if fname not in PROTECTED_FILES:
                    broken_tool_files.add(fname)

            # Undefined handlers in registry.go
            match = re.search(r"undefined:\s*(Handle\w+)", line)
            if match:
                undefined_handlers.add(match.group(1))

        # Remove broken tool files
        for fname in broken_tool_files:
            full_path = TOOLS_DIR / fname
            if full_path.exists():
                log.warn(f"Removing broken: {fname}")
                full_path.unlink()

        # Clean registry references
        if undefined_handlers:
            _clean_registry_refs(undefined_handlers)

        with log.lock:
            log.stats["builds_broken"] += 1

        return False, list(broken_tool_files), len(error_lines)

    except subprocess.TimeoutExpired:
        log.error("Build timed out!")
        return False, [], -1
    except Exception as e:
        log.error(f"Build error: {e}")
        return False, [], -1


def _clean_registry_refs(handlers_to_remove: set):
    """Remove handler references from registry.go for deleted files."""
    if not handlers_to_remove:
        return
    try:
        with open(REGISTRY_FILE, "r", encoding="utf-8", errors="replace") as f:
            lines = f.readlines()

        new_lines = []
        removed = 0
        for line in lines:
            if any(h in line for h in handlers_to_remove):
                removed += 1
                continue
            new_lines.append(line)

        with open(REGISTRY_FILE, "w", encoding="utf-8") as f:
            f.writelines(new_lines)

        if removed:
            log.info(f"Cleaned {removed} undefined handler refs from registry.go")
    except Exception as e:
        log.error(f"Failed to clean registry: {e}")


def repair_build(max_iterations=10):
    """Iteratively fix build errors until clean."""
    log.status("Repairing build...")
    for i in range(max_iterations):
        success, broken, errors = verify_build()
        if success:
            log.success(f"Build repaired after {i} iterations")
            return True
        if not broken and errors > 0:
            # Errors in non-tool files — can't auto-fix, break
            log.warn("Errors in non-tool files, can't auto-fix")
            break
        if errors == -1:
            log.error("Build process failed, waiting...")
            time.sleep(10)

    # Final check
    success, _, _ = verify_build()
    return success


# ══════════════════════════════════════════════════════════════════════════════
# MANIFEST MERGING
# ══════════════════════════════════════════════════════════════════════════════

_registry_merge_lock = threading.Lock()


def merge_manifests_into_registry():
    """Atomically merge all pending manifest files into registry.go."""
    with _registry_merge_lock:
        if not MANIFEST_DIR.exists():
            return 0

        manifests = list(MANIFEST_DIR.glob("*.json"))
        if not manifests:
            return 0

        all_handlers = []
        for mf in manifests:
            try:
                with open(mf, "r", encoding="utf-8") as f:
                    data = json.load(f)
                for h in data.get("handlers", []):
                    tool_name = h.get("tool_name", "")
                    handler_func = h.get("handler_func", "")
                    if tool_name and handler_func:
                        all_handlers.append((tool_name, handler_func))
                mf.unlink()
            except Exception as e:
                log.warn(f"Bad manifest {mf.name}: {e}")
                mf.unlink()

        if not all_handlers:
            return 0

        try:
            with open(REGISTRY_FILE, "r", encoding="utf-8", errors="replace") as f:
                content = f.read()
        except Exception as e:
            log.error(f"Cannot read registry.go: {e}")
            return 0

        existing_tools = set(re.findall(r'r\.handlers\["([^"]+)"\]', content))

        new_lines = []
        added = 0
        for tool_name, handler_func in all_handlers:
            if tool_name not in existing_tools and handler_func not in content:
                new_lines.append(f'\tr.handlers["{tool_name}"] = {handler_func}')
                added += 1
                existing_tools.add(tool_name)

        if added == 0:
            return 0

        insert_text = "\n".join(new_lines) + "\n"

        # Insert before closing brace of registerAll()
        lines = content.split("\n")
        insert_idx = -1
        brace_depth = 0
        in_func = False
        for i, line in enumerate(lines):
            if "func (r *Registry) registerAll()" in line:
                in_func = True
                brace_depth = 0
            if in_func:
                brace_depth += line.count("{") - line.count("}")
                if brace_depth <= 0 and i > 0:
                    insert_idx = i
                    in_func = False
                    break

        if insert_idx < 0:
            for i in range(len(lines) - 1, -1, -1):
                if lines[i].strip() == "}":
                    insert_idx = i
                    break

        if insert_idx >= 0:
            lines.insert(insert_idx, insert_text)
            content = "\n".join(lines)
            with open(REGISTRY_FILE, "w", encoding="utf-8") as f:
                f.write(content)
            log.success(f"Merged {added} new handler registrations")

        return added


# ══════════════════════════════════════════════════════════════════════════════
# WORKER — PROCESS A SINGLE TASK
# ══════════════════════════════════════════════════════════════════════════════


def process_task(task: dict, worker_id: int) -> bool:
    """Process a single MCP server task. Returns True on success."""
    name = task["name"]
    go_filename = server_name_to_go_filename(name)

    log.info(f"Processing: {name} -> {go_filename}.go", worker_id)

    # Choose model based on worker and task
    model_idx = (worker_id * 3 + hash(name)) % len(ALL_MODELS)
    model = ALL_MODELS[model_idx]

    # Build and call LLM
    prompt, system = build_prompt(task)
    output = llm.call(prompt, system=system, model=model, worker_id=worker_id)

    if not output:
        mark_server_failed(name, "LLM call failed after retries")
        log.error(f"LLM failed for {name}", worker_id)
        log.inc_failed()
        return False

    # Parse
    go_code, manifest = parse_llm_output(output, go_filename)

    if not go_code:
        mark_server_failed(name, "No Go code in LLM output")
        log.error(f"No Go code for {name}", worker_id)
        log.inc_failed()
        return False

    # Validate
    issues = validate_go_code(go_code)
    if issues:
        log.warn(f"Code issues: {'; '.join(issues[:3])}", worker_id)

    # Write files
    go_file_path = f"{go_filename}.go"
    if not write_go_file(go_code, go_file_path):
        mark_server_failed(name, "Failed to write Go file")
        log.inc_failed()
        return False

    if manifest:
        write_manifest(manifest, go_filename)

    # Update DB
    tools_count = len(manifest.get("handlers", [])) if manifest else 0
    mark_server_implemented(name, go_file_path, tools_count)

    log.success(f"Done: {name} ({tools_count} tools) -> {go_file_path}", worker_id)
    log.inc_completed()
    return True


# ══════════════════════════════════════════════════════════════════════════════
# ORCHESTRATOR
# ══════════════════════════════════════════════════════════════════════════════


class SwarmOrchestrator:
    """Main orchestrator — manages workers, batches, and build verification."""

    def __init__(self, num_workers=2, task_limit=50, forever=False, repair_only=False):
        self.num_workers = num_workers
        self.task_limit = task_limit
        self.forever = forever
        self.repair_only = repair_only
        self.running = True
        self.tasks_completed = 0
        self._shutdown_event = threading.Event()
        signal.signal(signal.SIGINT, self._signal_handler)
        signal.signal(signal.SIGTERM, self._signal_handler)

    def _signal_handler(self, signum, frame):
        log.status(f"Signal {signum} — shutting down gracefully...")
        self.running = False
        self._shutdown_event.set()

    def run(self):
        """Main entry point."""
        log.status("=" * 70)
        log.status("  TORMENTNEXUS SWARM v4 — THE INFINITE ASSIMILATOR")
        log.status("=" * 70)
        log.status(
            f"Workers: {self.num_workers} | Limit: {self.task_limit} | Forever: {self.forever}"
        )
        log.status(f"Proxy: {PROXY_URL} | Workspace: {WORKSPACE}")
        log.status(f"Models: {len(ALL_MODELS)} available")

        total, impl, pend, fail = get_db_stats()
        log.status(
            f"DB: {total} total, {impl} implemented, {pend} pending, {fail} failed"
        )
        log.status("=" * 70)

        MANIFEST_DIR.mkdir(exist_ok=True)

        # Check proxy
        if not proxy_mgr.check_health():
            log.warn("Proxy not healthy — attempting restart...")
            if not proxy_mgr.restart():
                log.error("Cannot start proxy — aborting!")
                return False

        # Repair-only mode
        if self.repair_only:
            repair_build()
            merge_manifests_into_registry()
            verify_build()
            return True

        # Initial build repair
        log.status("Checking build state...")
        repair_build()
        merge_manifests_into_registry()
        verify_build()

        # Main loop
        while self.running:
            try:
                batch_ok = self._run_batch()
                if not batch_ok and not self.forever:
                    break
                if not self.forever and self.tasks_completed >= self.task_limit:
                    log.status(f"Task limit ({self.task_limit}) reached!")
                    break
                if self.forever:
                    # Reset failed tasks for retry
                    self._reset_failed_for_retry()
                    time.sleep(10)
            except Exception as e:
                log.error(f"Batch error: {e}")
                traceback.print_exc()
                if not self.forever:
                    break
                time.sleep(30)

        self._shutdown()
        return True

    def _run_batch(self) -> bool:
        """Run a batch of tasks."""
        remaining = self.task_limit - self.tasks_completed
        pending = get_pending_tasks(limit=max(remaining, 20))
        retries = get_retry_tasks(limit=5) if remaining > 5 else []
        tasks = pending + retries

        if not tasks:
            if self.forever:
                log.status("No pending tasks — waiting 60s...")
                time.sleep(60)
                return True
            log.status("No pending tasks!")
            return False

        log.status(
            f"Batch: {len(tasks)} tasks ({len(pending)} new, {len(retries)} retries)"
        )

        batch_ok = 0
        batch_fail = 0

        # Process tasks ONE AT A TIME per worker (serial within each)
        # This prevents proxy saturation
        with ThreadPoolExecutor(max_workers=self.num_workers) as executor:
            futures = {}
            for i, task in enumerate(tasks[: self.task_limit]):
                if not self.running:
                    break
                worker_id = (i % self.num_workers) + 1
                future = executor.submit(self._worker_with_retry, task, worker_id)
                futures[future] = (task, worker_id)

            for future in as_completed(futures):
                if not self.running:
                    break
                task, wid = futures[future]
                try:
                    result = future.result()
                    if result:
                        self.tasks_completed += 1
                        batch_ok += 1
                    else:
                        batch_fail += 1
                except Exception as e:
                    log.error(f"Worker exception for {task['name']}: {e}", wid)
                    batch_fail += 1

                # Periodic build verification
                if (batch_ok + batch_fail) % BUILD_BATCH_SIZE == 0:
                    self._verify_and_merge()

        # Final verification for batch
        self._verify_and_merge()

        total, impl, pend, fail = get_db_stats()
        log.status(
            f"Batch done: {batch_ok} ok, {batch_fail} fail | "
            f"DB: {impl}/{total} implemented, {pend} pending"
        )
        log.status(log.stats_summary())
        return True

    def _worker_with_retry(self, task, worker_id):
        """Worker with retry logic around a single task."""
        for attempt in range(3):
            if not self.running:
                return False

            try:
                success = process_task(task, worker_id)
                if success:
                    return True

                # Circuit breaker open? Wait
                if llm.circuit_open:
                    wait = min(CIRCUIT_BREAKER_RESET, 120)
                    log.warn(f"Circuit open — waiting {wait}s", worker_id)
                    self._shutdown_event.wait(wait)

                # Check proxy
                if not proxy_mgr.is_available():
                    log.warn("Proxy down — waiting...", worker_id)
                    for _ in range(12):
                        if self._shutdown_event.is_set():
                            return False
                        if proxy_mgr.check_health():
                            break
                        time.sleep(5)
                    if not proxy_mgr.is_available():
                        proxy_mgr.restart()

                time.sleep(10 * (attempt + 1))

            except Exception as e:
                log.error(f"Worker error (attempt {attempt + 1}): {e}", worker_id)
                traceback.print_exc()
                time.sleep(30)

        return False

    def _verify_and_merge(self):
        """Merge manifests and verify build."""
        merge_manifests_into_registry()
        repair_build()

    def _reset_failed_for_retry(self):
        """Reset some failed tasks back to pending for retry."""
        db = sqlite3.connect(str(DB_PATH))
        c = db.cursor()
        # Reset the oldest failed tasks
        c.execute("""
            UPDATE mcp_servers SET status='pending', notes='auto-retry',
            updated_at=CURRENT_TIMESTAMP
            WHERE status='failed' AND score >= 80
            LIMIT 10
        """)
        count = c.rowcount
        db.commit()
        db.close()
        if count > 0:
            log.info(f"Reset {count} failed tasks for retry")

    def _shutdown(self):
        """Graceful shutdown."""
        log.status("Shutting down — merging manifests and verifying build...")
        merge_manifests_into_registry()
        repair_build()

        tool_count = len(list(TOOLS_DIR.glob("*.go")))
        try:
            handler_count = len(
                re.findall(
                    r'r\.handlers\["',
                    REGISTRY_FILE.read_text(encoding="utf-8", errors="replace"),
                )
            )
        except Exception:
            handler_count = 0

        total, impl, pend, fail = get_db_stats()

        log.status("=" * 70)
        log.status("  FINAL STATE")
        log.status("=" * 70)
        log.status(f"  Tool files: {tool_count}")
        log.status(f"  Registered handlers: {handler_count}")
        log.status(f"  DB: {impl}/{total} implemented, {pend} pending, {fail} failed")
        log.status(f"  {log.stats_summary()}")
        log.status("=" * 70)


# ══════════════════════════════════════════════════════════════════════════════
# MAIN
# ══════════════════════════════════════════════════════════════════════════════


def main():
    parser = argparse.ArgumentParser(description="TormentNexus Swarm v4")
    parser.add_argument("--workers", type=int, default=2, help="Concurrent workers")
    parser.add_argument("--limit", type=int, default=50, help="Max tasks")
    parser.add_argument("--forever", action="store_true", help="Run continuously")
    parser.add_argument("--repair", action="store_true", help="Repair build only")
    parser.add_argument("--proxy", type=str, default=None, help="Proxy URL")
    parser.add_argument("--log", type=str, default=None, help="Log file")

    args = parser.parse_args()

    if args.proxy:
        global PROXY_URL
        PROXY_URL = args.proxy

    if args.log:
        log.log_file = args.log
    else:
        log.log_file = str(
            WORKSPACE
            / "data"
            / f"swarm_v4_{datetime.now().strftime('%Y%m%d_%H%M%S')}.log"
        )

    orchestrator = SwarmOrchestrator(
        num_workers=args.workers,
        task_limit=args.limit,
        forever=args.forever,
        repair_only=args.repair,
    )

    success = orchestrator.run()
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
