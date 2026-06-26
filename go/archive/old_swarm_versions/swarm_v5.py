#!/usr/bin/env python3
"""
TormentNexus Swarm v5 — THE COORDINATED ASSIMILATOR
====================================================

Multi-agent system that assimilates MCP servers into native Go modules.

ARCHITECTURE:
- Workers CLAIM tasks atomically from SQLite (no double-processing)
- LLM calls go DIRECTLY to NVIDIA NIM (bypasses proxy 30s timeout)
- Falls back to FreeLLM proxy for models not on NVIDIA
- Workers write .go files + manifests
- Single-threaded MERGER integrates manifests into registry.go
- BUILD VERIFIER removes broken files, fixes registry refs
- Auto-retries failed tasks, auto-restarts crashed proxy

COORDINATION:
  SQLite DB (single source of truth)
       │
  ┌────┼────┐
  │    │    │
  W1   W2   W8    ← Each CLAIMS a unique server
  │    │    │
  └────┼────┘
       │
   MERGER → registry.go
       │
   BUILD CHECK

USAGE:
  python3 swarm_v5.py --workers 8 --limit 50
  python3 swarm_v5.py --workers 8 --forever
  python3 swarm_v5.py --repair
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

# ══════════════════════════════════════════════════════════════════════════════
# CONFIGURATION
# ══════════════════════════════════════════════════════════════════════════════

PROXY_URL = os.environ.get("SWARM_PROXY", "http://localhost:4000")
PROXY_KEY = os.environ.get("SWARM_KEY", "sk-freellm")
PROXY_BIN = os.environ.get(
    "SWARM_PROXY_BIN",
    "C:/Users/hyper/workspace/litellm_control_panel/freellm.exe",
)
NVIDIA_KEY = os.environ.get("NVIDIA_API_KEY", "")
NVIDIA_BASE = "https://integrate.api.nvidia.com/v1"

WORKSPACE = Path(
    os.environ.get("SWARM_WORKSPACE", "C:/Users/hyper/workspace/tormentnexus")
)
GO_DIR = WORKSPACE / "go"
TOOLS_DIR = GO_DIR / "internal" / "tools"
REGISTRY_FILE = TOOLS_DIR / "registry.go"
DB_PATH = WORKSPACE / "data" / "assimilation_state.db"
MANIFEST_DIR = TOOLS_DIR / "manifests"

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

# NVIDIA NIM models (direct, no proxy timeout)
NVIDIA_MODELS = [
    "qwen/qwen3.5-397b-a17b",  # Best quality, ~70s per code gen
    "qwen/qwen3-coder-480b-a35b-instruct",  # Code specialist
    "deepseek-ai/deepseek-v4-flash",  # Fast deepseek
    "nvidia/llama-3.1-nemotron-ultra-253b-v1",  # Largest
    "mistralai/mistral-large-3-675b-instruct-2512",
    "meta/llama-3.3-70b-instruct",
]

# Proxy models (fallback, limited to 30s timeout = short output only)
PROXY_MODELS = [
    "gpt-4o-mini",
    "free-llm",
    "google/gemini-2.5-flash-lite",
]

MAX_RETRIES = 3
RETRY_DELAY = 30
CIRCUIT_BREAKER_TRIPS = 5
CIRCUIT_BREAKER_COOLDOWN = 120
PROXY_RESTART_AFTER = 8
BUILD_VERIFY_EVERY = 3


# ══════════════════════════════════════════════════════════════════════════════
# LOGGING
# ══════════════════════════════════════════════════════════════════════════════


class Log:
    C = {
        "R": "\033[91m",
        "G": "\033[92m",
        "Y": "\033[93m",
        "B": "\033[94m",
        "M": "\033[95m",
        "C": "\033[96m",
        "0": "\033[0m",
    }

    def __init__(self, path=None):
        self.lock = threading.Lock()
        self.path = path
        self.t0 = time.time()
        self.ok_count = 0
        self.fail_count = 0
        self.llm_ok_count = 0
        self.llm_err_count = 0
        self.proxy_restarts = 0

    def _ts(self):
        e = int(time.time() - self.t0)
        return f"{e // 3600:02d}:{(e % 3600) // 60:02d}:{e % 60:02d}"

    def _w(self, lvl, col, wid, msg):
        p = f"[{self._ts()}]"
        if wid is not None:
            p += f"[W{wid}]"
        p += f"[{lvl}] {msg}"
        with self.lock:
            print(f"{self.C.get(col, '')}{p}{self.C['0']}", flush=True)
            if self.path:
                try:
                    with open(self.path, "a", encoding="utf-8", errors="replace") as f:
                        f.write(p + "\n")
                except Exception:
                    pass

    def info(self, m, w=None):
        self._w("INFO", "C", w, m)

    def ok(self, m, w=None):
        self._w("OK", "G", w, m)
        with self.lock:
            self.ok_count += 1

    def warn(self, m, w=None):
        self._w("WARN", "Y", w, m)

    def err(self, m, w=None):
        self._w("ERR", "R", w, m)
        with self.lock:
            self.fail_count += 1

    def llm(self, m, w=None):
        self._w("LLM", "M", w, m)
        with self.lock:
            self.llm_ok_count += 1

    def llm_err(self, m, w=None):
        self._w("LLM-ERR", "R", w, m)
        with self.lock:
            self.llm_err_count += 1

    def stat(self, m):
        self._w("STAT", "B", None, m)

    def summary(self):
        e = int(time.time() - self.t0)
        with self.lock:
            return (
                f"ok={self.ok_count} fail={self.fail_count} "
                f"llm_ok={self.llm_ok_count} llm_err={self.llm_err_count} "
                f"restarts={self.proxy_restarts} "
                f"elapsed={e // 3600}h{(e % 3600) // 60}m"
            )


log = Log()


# ══════════════════════════════════════════════════════════════════════════════
# LLM CLIENT — Direct NVIDIA NIM + Proxy fallback
# ══════════════════════════════════════════════════════════════════════════════


class LLMClient:
    def __init__(self):
        self.lock = threading.Lock()
        self.nvidia_idx = 0
        self.proxy_idx = 0
        self.circuit_open = False
        self.circuit_opened_at = 0
        self.consecutive_fails = 0
        self.proxy_consecutive_errs = 0

    def _next_nvidia(self):
        with self.lock:
            m = NVIDIA_MODELS[self.nvidia_idx % len(NVIDIA_MODELS)]
            self.nvidia_idx += 1
            return m

    def _next_proxy(self):
        with self.lock:
            m = PROXY_MODELS[self.proxy_idx % len(PROXY_MODELS)]
            self.proxy_idx += 1
            return m

    def _check_circuit(self):
        if not self.circuit_open:
            return True
        if time.time() - self.circuit_opened_at > CIRCUIT_BREAKER_COOLDOWN:
            log.stat("Circuit breaker reset")
            self.circuit_open = False
            self.consecutive_fails = 0
            return True
        return False

    def _trip_circuit(self):
        self.circuit_open = True
        self.circuit_opened_at = time.time()
        log.warn(f"Circuit OPEN — {CIRCUIT_BREAKER_COOLDOWN}s cooldown")

    def _record_fail(self):
        self.consecutive_fails += 1
        if self.consecutive_fails >= CIRCUIT_BREAKER_TRIPS:
            self._trip_circuit()

    def _record_ok(self):
        self.consecutive_fails = 0

    def call_nvidia(self, prompt, system, model, timeout=300):
        """Call NVIDIA NIM directly."""
        import urllib.request

        payload = json.dumps(
            {
                "model": model,
                "messages": [
                    {"role": "system", "content": system},
                    {"role": "user", "content": prompt},
                ],
                "temperature": 0.3,
                "max_tokens": 4096,
            }
        ).encode("utf-8")
        req = urllib.request.Request(
            f"{NVIDIA_BASE}/chat/completions",
            data=payload,
            headers={
                "Content-Type": "application/json",
                "Authorization": f"Bearer {NVIDIA_KEY}",
            },
            method="POST",
        )
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            return data["choices"][0]["message"]["content"]

    def call_proxy(self, prompt, system, model, timeout=60):
        """Call FreeLLM proxy."""
        import urllib.request

        payload = json.dumps(
            {
                "model": model,
                "messages": [
                    {"role": "system", "content": system},
                    {"role": "user", "content": prompt},
                ],
                "temperature": 0.3,
                "max_tokens": 2048,
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
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            return data["choices"][0]["message"]["content"]

    def call(self, prompt, system="", wid=0):
        """Call LLM with retry. Tries NVIDIA first, proxy as fallback."""
        if not self._check_circuit():
            return None

        # Try NVIDIA NIM models (direct, no timeout issues)
        if NVIDIA_KEY:
            for attempt in range(MAX_RETRIES):
                model = self._next_nvidia()
                timeout = 120 + (attempt * 60)
                log.llm(
                    f"NVIDIA {model} attempt={attempt + 1}/{MAX_RETRIES} t={timeout}s",
                    wid,
                )
                try:
                    content = self.call_nvidia(prompt, system, model, timeout)
                    self._record_ok()
                    return content
                except Exception as e:
                    es = str(e)[:120]
                    log.llm_err(f"NVIDIA {model}: {es}", wid)
                    self._record_fail()
                    time.sleep(RETRY_DELAY * (attempt + 1))

        # Fallback: try proxy models (limited by 30s timeout)
        for attempt in range(2):
            model = self._next_proxy()
            log.llm(f"PROXY {model} attempt={attempt + 1}/2 t=55s", wid)
            try:
                content = self.call_proxy(prompt, system, model, timeout=55)
                self._record_ok()
                self.proxy_consecutive_errs = 0
                return content
            except Exception as e:
                es = str(e)[:120]
                log.llm_err(f"PROXY {model}: {es}", wid)
                self.proxy_consecutive_errs += 1
                self._record_fail()
                time.sleep(RETRY_DELAY)

        return None


llm = LLMClient()


# ══════════════════════════════════════════════════════════════════════════════
# PROXY MANAGER (for restart only — LLM calls go to NVIDIA now)
# ══════════════════════════════════════════════════════════════════════════════


class ProxyManager:
    def __init__(self):
        self._healthy = True

    def check(self):
        try:
            import urllib.request

            with urllib.request.urlopen(f"{PROXY_URL}/health", timeout=10) as r:
                if r.status == 200:
                    self._healthy = True
                    return True
        except Exception:
            pass
        self._healthy = False
        return False

    def restart(self):
        log.stat("Restarting proxy...")
        try:
            subprocess.run(
                ["taskkill", "//F", "//IM", "freellm.exe"],
                capture_output=True,
                timeout=10,
            )
        except Exception:
            pass
        time.sleep(15)
        try:
            d = str(Path(PROXY_BIN).parent)
            subprocess.Popen(
                [PROXY_BIN], cwd=d, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
            )
        except Exception as e:
            log.err(f"Proxy start: {e}")
            return False
        for i in range(24):
            time.sleep(5)
            if self.check():
                log.ok("Proxy back up!")
                with log.lock:
                    log.proxy_restarts += 1
                return True
        log.err("Proxy never came up")
        return False

    @property
    def healthy(self):
        return self._healthy


proxy = ProxyManager()


# ══════════════════════════════════════════════════════════════════════════════
# COORDINATION — Atomic task claim
# ══════════════════════════════════════════════════════════════════════════════

_db_lock = threading.Lock()


def claim_task(worker_id):
    """Atomically claim a pending server."""
    with _db_lock:
        db = sqlite3.connect(str(DB_PATH))
        db.isolation_level = "EXCLUSIVE"
        c = db.cursor()
        row = c.execute(
            "SELECT name, github_url, score, category, description, "
            "subcategory, tags FROM mcp_servers "
            "WHERE status='pending' ORDER BY score DESC LIMIT 1"
        ).fetchone()
        if not row:
            db.close()
            return None
        name = row[0]
        c.execute(
            "UPDATE mcp_servers SET status='processing', "
            "notes=?, updated_at=CURRENT_TIMESTAMP WHERE name=? AND status='pending'",
            (f"w{worker_id}", name),
        )
        db.commit()
        if c.rowcount == 0:
            db.close()
            return None
        db.close()
        return {
            "name": name,
            "github_url": row[1] or "",
            "score": row[2] or 0,
            "category": row[3] or "unknown",
            "description": row[4] or "",
            "subcategory": row[5] or "",
            "tags": row[6] or "",
        }


def complete_task(name, go_file, tools_count):
    with _db_lock:
        db = sqlite3.connect(str(DB_PATH))
        c = db.cursor()
        c.execute(
            "UPDATE mcp_servers SET status='implemented', go_file=?, "
            "tools_exposed=?, notes='v5', updated_at=CURRENT_TIMESTAMP WHERE name=?",
            (go_file, str(tools_count), name),
        )
        db.commit()
        db.close()


def fail_task(name, error):
    with _db_lock:
        db = sqlite3.connect(str(DB_PATH))
        c = db.cursor()
        c.execute(
            "UPDATE mcp_servers SET status='failed', notes=?, "
            "updated_at=CURRENT_TIMESTAMP WHERE name=?",
            (error[:500], name),
        )
        db.commit()
        db.close()


def release_task(name):
    with _db_lock:
        db = sqlite3.connect(str(DB_PATH))
        c = db.cursor()
        c.execute(
            "UPDATE mcp_servers SET status='pending', notes='released' "
            "WHERE name=? AND status='processing'",
            (name,),
        )
        db.commit()
        db.close()


def reset_stale():
    db = sqlite3.connect(str(DB_PATH))
    c = db.cursor()
    c.execute(
        "UPDATE mcp_servers SET status='pending', notes='stale' "
        "WHERE status='processing'"
    )
    n = c.rowcount
    db.commit()
    db.close()
    return n


def reset_failed(limit=30):
    db = sqlite3.connect(str(DB_PATH))
    c = db.cursor()
    c.execute(
        "UPDATE mcp_servers SET status='pending', notes='retry' "
        "WHERE status='failed' AND score >= 50 ORDER BY score DESC LIMIT ?",
        (limit,),
    )
    n = c.rowcount
    db.commit()
    db.close()
    return n


def get_stats():
    db = sqlite3.connect(str(DB_PATH))
    c = db.cursor()
    t = c.execute("SELECT COUNT(*) FROM mcp_servers").fetchone()[0]
    i = c.execute(
        'SELECT COUNT(*) FROM mcp_servers WHERE status="implemented"'
    ).fetchone()[0]
    p = c.execute('SELECT COUNT(*) FROM mcp_servers WHERE status="pending"').fetchone()[
        0
    ]
    f = c.execute('SELECT COUNT(*) FROM mcp_servers WHERE status="failed"').fetchone()[
        0
    ]
    db.close()
    return t, i, p, f


# ══════════════════════════════════════════════════════════════════════════════
# GO CODE GENERATION
# ══════════════════════════════════════════════════════════════════════════════


def name_to_filename(name):
    n = name
    for p in ["mcp-server-", "mcp-", "@", "server-", "mcp_"]:
        if n.lower().startswith(p):
            n = n[len(p) :]
    n = re.sub(r"[/.\-]+", "_", n)
    n = re.sub(r"[^a-zA-Z0-9_]", "", n).lower()
    if len(n) < 2:
        n = f"tool_{abs(hash(name)) % 10000}"
    return n[:60]


def make_prompt(task):
    n = task["name"]
    u = task.get("github_url", "")
    d = task.get("description", "")
    cat = task.get("category", "")
    fn = name_to_filename(n)
    prompt = textwrap.dedent(f"""\
Implement a Go-native MCP tool module for "{n}".
GitHub: {u} | Category: {cat} | Description: {d}

RULES:
1. Package: `package tools`
2. Handler: `func HandleXxx(ctx context.Context, args map[string]interface{{}}) (ToolResponse, error)`
3. Success: `return ok("text")` — err() returns (ToolResponse, error)
4. Error CHECK: `if e != nil {{ return err(e.Error()) }}`
5. getString/getInt/getBool return SINGLE values: `val := getString(args, "key")`
6. Only stdlib imports: context, encoding/json, fmt, io, net/http, net/url, os, os/exec, path/filepath, strconv, strings, time, regexp, sort
7. http.Client{{Timeout: 30*time.Second}}
8. 2-6 handlers, keep simple, MUST COMPILE, no TODOs

OUTPUT:
===GO_FILE===
(code)
===MANIFEST===
{{"filename":"{fn}.go","server_name":"{n}","handlers":[{{"tool_name":"x","handler_func":"HandleX","description":"d"}}]}}""")
    system = "You write clean compilable Go. ok()/err() return (ToolResponse,error). getString returns single value. No pseudocode."
    return prompt, system


def parse_output(output, go_filename):
    if not output:
        return None, None
    go_code = manifest = None
    gm = re.search(r"===GO_FILE===\s*\n(.*?)(?====MANIFEST===|$)", output, re.DOTALL)
    mm = re.search(r"===MANIFEST===\s*\n(.*?)$", output, re.DOTALL)
    if gm:
        go_code = re.sub(r"^```go\s*\n?|\n?```\s*$", "", gm.group(1).strip())
    if mm:
        ms = re.sub(r"^```json\s*\n?|\n?```\s*$", "", mm.group(1).strip())
        try:
            manifest = json.loads(ms)
        except Exception:
            ms = re.sub(r",\s*[}\]]", lambda m: m.group(0)[-1:], ms)
            try:
                manifest = json.loads(ms)
            except:
                pass
    if not go_code:
        cm = re.search(r"```go\s*\n(.*?)```", output, re.DOTALL)
        if cm:
            go_code = cm.group(1).strip()
    if not go_code:
        pm = re.search(r"(package tools\s+.*)", output, re.DOTALL)
        if pm:
            go_code = pm.group(1).strip()
    if go_code and not go_code.strip().startswith("package tools"):
        idx = go_code.find("package tools")
        go_code = go_code[idx:] if idx >= 0 else None
    if go_code and not manifest:
        handlers = re.findall(r"func (Handle\w+)\s*\(", go_code)
        if handlers:
            manifest = {
                "filename": f"{go_filename}.go",
                "handlers": [
                    {
                        "tool_name": re.sub(r"([A-Z])", r"_\1", h[6:])
                        .lower()
                        .lstrip("_"),
                        "handler_func": h,
                        "description": h,
                    }
                    for h in handlers
                ],
            }
    return go_code, manifest


def write_go(code, filename):
    try:
        with open(TOOLS_DIR / filename, "w", encoding="utf-8") as f:
            f.write(code)
        return True
    except Exception as e:
        log.err(f"Write {filename}: {e}")
        return False


def write_manifest(m, filename):
    MANIFEST_DIR.mkdir(exist_ok=True)
    try:
        with open(MANIFEST_DIR / f"{filename}.json", "w", encoding="utf-8") as f:
            json.dump(m, f, indent=2)
        return True
    except Exception:
        return False


# ══════════════════════════════════════════════════════════════════════════════
# BUILD VERIFICATION
# ══════════════════════════════════════════════════════════════════════════════


def verify_build():
    try:
        r = subprocess.run(
            ["go", "build", "-buildvcs=false", "./cmd/tormentnexus"],
            capture_output=True,
            text=True,
            cwd=str(GO_DIR),
            timeout=120,
            encoding="utf-8",
            errors="replace",
        )
        if r.returncode == 0:
            return True
        broken = set()
        undefs = set()
        for line in (r.stderr or "").split("\n"):
            m = re.match(r"internal[\\/]tools[\\/](\w[\w_]*\.go)", line)
            if m and m.group(1) not in PROTECTED_FILES:
                broken.add(m.group(1))
            m = re.search(r"undefined:\s*(Handle\w+)", line)
            if m:
                undefs.add(m.group(1))
        for f in broken:
            p = TOOLS_DIR / f
            if p.exists():
                log.warn(f"Removing broken: {f}")
                p.unlink()
        if undefs:
            _clean_registry(undefs)
        return False
    except Exception as e:
        log.err(f"Build: {e}")
        return False


def _clean_registry(handlers):
    try:
        with open(REGISTRY_FILE, "r", encoding="utf-8", errors="replace") as f:
            lines = f.readlines()
        new = [l for l in lines if not any(h in l for h in handlers)]
        removed = len(lines) - len(new)
        with open(REGISTRY_FILE, "w", encoding="utf-8") as f:
            f.writelines(new)
        if removed:
            log.info(f"Cleaned {removed} undefined refs")
    except Exception as e:
        log.err(f"Registry: {e}")


def repair_build(max_iter=10):
    for i in range(max_iter):
        if verify_build():
            log.ok(f"Build clean (iter {i})")
            return True
    return verify_build()


# ══════════════════════════════════════════════════════════════════════════════
# MANIFEST MERGER
# ══════════════════════════════════════════════════════════════════════════════

_merge_lock = threading.Lock()


def merge_manifests():
    with _merge_lock:
        if not MANIFEST_DIR.exists():
            return 0
        mfs = list(MANIFEST_DIR.glob("*.json"))
        if not mfs:
            return 0
        all_h = []
        for mf in mfs:
            try:
                with open(mf, "r", encoding="utf-8") as f:
                    data = json.load(f)
                for h in data.get("handlers", []):
                    tn, hf = h.get("tool_name", ""), h.get("handler_func", "")
                    if tn and hf:
                        all_h.append((tn, hf))
                mf.unlink()
            except Exception:
                mf.unlink()
        if not all_h:
            return 0
        try:
            with open(REGISTRY_FILE, "r", encoding="utf-8", errors="replace") as f:
                content = f.read()
        except Exception:
            return 0
        existing = set(re.findall(r'r\.handlers\["([^"]+)"\]', content))
        to_add = []
        added = 0
        for tn, hf in all_h:
            if tn not in existing and hf not in content:
                to_add.append(f'\tr.handlers["{tn}"] = {hf}')
                added += 1
                existing.add(tn)
        if not added:
            return 0
        insert = "\n".join(to_add) + "\n"
        lines = content.split("\n")
        idx = -1
        depth, in_func = 0, False
        for i, line in enumerate(lines):
            if "func (r *Registry) registerAll()" in line:
                in_func, depth = True, 0
            if in_func:
                depth += line.count("{") - line.count("}")
                if depth <= 0 and i > 0:
                    idx = i
                    in_func = False
                    break
        if idx < 0:
            for i in range(len(lines) - 1, -1, -1):
                if lines[i].strip() == "}":
                    idx = i
                    break
        if idx >= 0:
            lines.insert(idx, insert)
            with open(REGISTRY_FILE, "w", encoding="utf-8") as f:
                f.write("\n".join(lines))
            log.ok(f"Merged {added} handlers")
        return added


# ══════════════════════════════════════════════════════════════════════════════
# WORKER
# ══════════════════════════════════════════════════════════════════════════════


def worker_loop(worker_id, shutdown):
    """Main worker loop: claim → generate → write → repeat."""
    while not shutdown.is_set():
        task = claim_task(worker_id)
        if not task:
            return  # No more tasks

        name = task["name"]
        fn = name_to_filename(name)
        log.info(f"Claimed: {name} -> {fn}.go", worker_id)

        # Wait if circuit breaker is open
        if llm.circuit_open:
            log.warn("Circuit open — waiting...", worker_id)
            for _ in range(12):
                if shutdown.is_set():
                    release_task(name)
                    return
                if not llm.circuit_open:
                    break
                time.sleep(10)

        # Call LLM
        prompt, system = make_prompt(task)
        output = llm.call(prompt, system, wid=worker_id)

        if not output:
            fail_task(name, "LLM failed after retries")
            log.err(f"LLM failed: {name}", worker_id)
            continue

        # Parse + write
        go_code, manifest = parse_output(output, fn)
        if not go_code:
            fail_task(name, "No Go code in output")
            log.err(f"No code: {name}", worker_id)
            continue

        go_file = f"{fn}.go"
        if not write_go(go_code, go_file):
            fail_task(name, "Write failed")
            continue

        if manifest:
            write_manifest(manifest, fn)

        tc = len(manifest.get("handlers", [])) if manifest else 0
        complete_task(name, go_file, tc)
        log.ok(f"Done: {name} ({tc} tools)", worker_id)


# ══════════════════════════════════════════════════════════════════════════════
# ORCHESTRATOR
# ══════════════════════════════════════════════════════════════════════════════


class Swarm:
    def __init__(self, max_workers, limit, forever, repair_only):
        self.max_workers = max_workers
        self.limit = limit
        self.forever = forever
        self.repair_only = repair_only
        self.running = True
        self.completed = 0
        self.shutdown = threading.Event()
        signal.signal(signal.SIGINT, lambda s, f: self._stop())
        signal.signal(signal.SIGTERM, lambda s, f: self._stop())

    def _stop(self):
        log.stat("Shutdown signal...")
        self.running = False
        self.shutdown.set()

    def run(self):
        log.stat("=" * 60)
        log.stat("  TORMENTNEXUS SWARM v5 — COORDINATED ASSIMILATOR")
        log.stat("=" * 60)
        log.stat(
            f"Workers: {self.max_workers} | Limit: {self.limit} | Forever: {self.forever}"
        )
        log.stat(f"NVIDIA models: {NVIDIA_MODELS[:3]}... ({len(NVIDIA_MODELS)} total)")
        log.stat(f"NVIDIA_KEY: {'set' if NVIDIA_KEY else 'NOT SET'}")
        log.stat(f"Proxy: {PROXY_URL} (fallback)")

        t, i, p, f = get_stats()
        log.stat(f"DB: {t} total | {i} done | {p} pending | {f} failed")

        MANIFEST_DIR.mkdir(exist_ok=True)
        stale = reset_stale()
        if stale:
            log.info(f"Released {stale} stale tasks")

        if self.repair_only:
            repair_build()
            merge_manifests()
            verify_build()
            return True

        log.stat("Checking build...")
        repair_build()
        merge_manifests()
        verify_build()

        while self.running:
            try:
                self._run_wave()
            except Exception as e:
                log.err(f"Wave: {e}")
                traceback.print_exc()
                if not self.forever:
                    break
                time.sleep(30)

            if not self.forever and self.completed >= self.limit:
                log.stat(f"Limit reached ({self.limit})!")
                break

            _, _, pending, _ = get_stats()
            if pending == 0:
                if self.forever:
                    n = reset_failed(50)
                    if n == 0:
                        log.stat("ALL DONE!")
                        break
                    log.stat(f"Retrying {n} failed tasks")
                else:
                    log.stat("No pending!")
                    break

            if self.forever:
                time.sleep(5)

        self._shutdown()
        return True

    def _run_wave(self):
        ws = (
            min(self.max_workers, self.limit - self.completed)
            if not self.forever
            else self.max_workers
        )
        if ws <= 0:
            return
        log.stat(f"Wave: {ws} workers")

        before = self.completed
        bc = 0

        with ThreadPoolExecutor(max_workers=ws) as pool:
            futs = {}
            for wid in range(ws):
                if self.shutdown.is_set():
                    break
                futs[pool.submit(worker_loop, wid + 1, self.shutdown)] = wid + 1

            for f in as_completed(futs):
                if self.shutdown.is_set():
                    break
                try:
                    f.result()
                    self.completed += 1
                    bc += 1
                except Exception as e:
                    log.err(f"Worker {futs[f]}: {e}")

                if bc >= BUILD_VERIFY_EVERY:
                    merge_manifests()
                    repair_build()
                    bc = 0

        merge_manifests()
        repair_build()

        t, i, p, f2 = get_stats()
        log.stat(
            f"Wave: +{self.completed - before} | DB: {i}/{t} done, {p} pend | {log.summary()}"
        )

        if self.forever:
            n = reset_failed(20)
            if n:
                log.info(f"Retry: {n} failed")

    def _shutdown(self):
        log.stat("Shutting down...")
        merge_manifests()
        repair_build()
        tc = len(list(TOOLS_DIR.glob("*.go")))
        try:
            hc = len(
                re.findall(
                    r'r\.handlers\["',
                    REGISTRY_FILE.read_text(encoding="utf-8", errors="replace"),
                )
            )
        except Exception:
            hc = 0
        t, i, p, f = get_stats()
        log.stat("=" * 60)
        log.stat(f"  Files: {tc} | Handlers: {hc}")
        log.stat(f"  DB: {i}/{t} done, {p} pend, {f} fail")
        log.stat(f"  {log.summary()}")
        log.stat("=" * 60)


def main():
    ap = argparse.ArgumentParser(description="TormentNexus Swarm v5")
    ap.add_argument("--workers", type=int, default=8)
    ap.add_argument("--limit", type=int, default=50)
    ap.add_argument("--forever", action="store_true")
    ap.add_argument("--repair", action="store_true")
    ap.add_argument("--log", type=str, default=None)
    args = ap.parse_args()

    log.path = args.log or str(
        WORKSPACE / "data" / f"swarm_v5_{datetime.now().strftime('%Y%m%d_%H%M%S')}.log"
    )

    global NVIDIA_KEY
    NVIDIA_KEY = os.environ.get("NVIDIA_API_KEY", "")

    s = Swarm(args.workers, args.limit, args.forever, args.repair)
    sys.exit(0 if s.run() else 1)


if __name__ == "__main__":
    main()
