#!/usr/bin/env python3
"""
TormentNexus Swarm v2 — Aggressive parallel LLM workers via FreeLLM proxy.

Design principles:
  - Pre-research: README + tool specs embedded in task, no wasted turns
  - Immediate implementation: workers told to write code, not research
  - Continuous retry: failed tasks re-queued with error context added
  - Model rotation: cycle through 6+ free models to spread load
  - Atomic registry: file lock for registry.go, batch registration
  - Build verification: mandatory go build after every Go file write
  - Auto-fix: on build failure, feed error back to worker for repair

Usage:
  python3 swarm_v2.py --workers 8 --limit 50
  python3 swarm_v2.py --workers 4 --limit 20 --model free-llm
"""

import argparse
import json
import os
import re
import sqlite3
import subprocess
import sys
import threading
import time
import traceback
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor, as_completed

import requests

# Limit concurrent API calls to avoid overwhelming the proxy
_api_semaphore = threading.Semaphore(4)

# ─── Config ────────────────────────────────────────────────────────────────
PROXY = os.environ.get("SWARM_PROXY", "http://localhost:4000")
API_KEY = os.environ.get("SWARM_KEY", "sk-freellm")
WORKSPACE = os.environ.get("SWARM_WORKSPACE", "C:/Users/hyper/workspace/tormentnexus")
STATE_DB = f"{WORKSPACE}/data/assimilation_state.db"
GO_DIR = f"{WORKSPACE}/go/internal/tools"
REGISTRY = f"{GO_DIR}/registry.go"
BUILD_CMD = f"cd {WORKSPACE}/go && go build -buildvcs=false ./cmd/tormentnexus"

MODELS = [
    "deepseek/deepseek-v4-flash",
    "qwen/qwen3.6-flash",
    "gpt-4o-mini",
    "gemini-3.5-flash",
]

MAX_TURNS = 25
API_TIMEOUT = 150
RETRY_LIMIT = 5
CB_FAIL_THRESHOLD = 3
CB_RESET_SECS = 60


# ─── Tools available to LLM workers ────────────────────────────────────────
WORKER_TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "read_file",
            "description": "Read a file's contents.",
            "parameters": {
                "type": "object",
                "properties": {"path": {"type": "string", "description": "File path"}},
                "required": ["path"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "write_file",
            "description": "Write content to a file. Creates directories if needed.",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {"type": "string", "description": "File path"},
                    "content": {"type": "string", "description": "Content to write"},
                },
                "required": ["path", "content"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "bash",
            "description": "Run a shell command. Returns stdout+stderr.",
            "parameters": {
                "type": "object",
                "properties": {"cmd": {"type": "string", "description": "Command to run"}},
                "required": ["cmd"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "grep",
            "description": "Search for a pattern in Go files.",
            "parameters": {
                "type": "object",
                "properties": {
                    "pattern": {"type": "string", "description": "Search pattern"},
                    "dir": {"type": "string", "description": "Directory (default: go/internal/tools)"},
                },
                "required": ["pattern"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "done",
            "description": "Mark your task as complete. Provide a summary of what you implemented.",
            "parameters": {
                "type": "object",
                "properties": {
                    "summary": {"type": "string", "description": "What you implemented"},
                    "files": {
                        "type": "array",
                        "items": {"type": "string"},
                        "description": "Files you created/modified",
                    },
                },
                "required": ["summary"],
            },
        },
    },
]


def exec_tool(name: str, args: dict) -> str:
    """Execute a tool call and return the result."""
    try:
        if name == "read_file":
            p = Path(args["path"])
            if not p.exists():
                return f"ERROR: file not found: {p}"
            txt = p.read_text(encoding="utf-8", errors="replace")
            return txt[:50000] + (f"\n...[truncated {len(txt)}B]" if len(txt) > 50000 else "")

        elif name == "write_file":
            p = Path(args["path"])
            p.parent.mkdir(parents=True, exist_ok=True)
            p.write_text(args["content"], encoding="utf-8")
            return f"WROTE {len(args['content'])}B -> {p}"

        elif name == "bash":
            r = subprocess.run(
                args["cmd"], shell=True, capture_output=True, text=True,
                timeout=180, cwd=WORKSPACE, encoding="utf-8", errors="replace",
            )
            out = (r.stdout or "") + (f"\nSTDERR:\n{r.stderr}" if r.stderr else "")
            return (out[:30000] + "...[truncated]") if len(out) > 30000 else (out or f"(exit={r.returncode})")

        elif name == "grep":
            d = args.get("dir", GO_DIR)
            r = subprocess.run(
                ["grep", "-rn", "--include=*.go", "-E", args["pattern"], d],
                capture_output=True, text=True, timeout=30, encoding="utf-8", errors="replace",
            )
            return (r.stdout[:20000] or "(no matches)") + ("...[truncated]" if len(r.stdout) > 20000 else "")

        elif name == "done":
            return "__DONE__"

        else:
            return f"ERROR: unknown tool {name}"

    except subprocess.TimeoutExpired:
        return "ERROR: command timed out (180s)"
    except Exception as e:
        return f"ERROR: {type(e).__name__}: {str(e)[:500]}"


# ─── LLM client with circuit breaker ───────────────────────────────────────
class ModelPool:
    def __init__(self, models):
        self.models = models
        self._fails = {m: 0 for m in models}
        self._cb_until = {m: 0.0 for m in models}
        self._idx = 0
        self._lock = threading.Lock()

    def pick(self):
        with self._lock:
            now = time.time()
            for _ in range(len(self.models)):
                m = self.models[self._idx % len(self.models)]
                self._idx += 1
                if self._fails[m] >= CB_FAIL_THRESHOLD and now < self._cb_until.get(m, 0):
                    continue
                if self._fails[m] >= CB_FAIL_THRESHOLD:
                    self._fails[m] = 0  # reset circuit
                return m
            return self.models[0]  # fallback

    def ok(self, m):
        with self._lock:
            self._fails[m] = 0

    def fail(self, m):
        with self._lock:
            self._fails[m] += 1
            if self._fails[m] >= CB_FAIL_THRESHOLD:
                self._cb_until[m] = time.time() + CB_RESET_SECS


def llm_call(model, messages, pool=None):
    """Call LLM, return parsed response dict. Raises on HTTP errors."""
    with _api_semaphore:
        r = requests.post(
        f"{PROXY}/v1/chat/completions",
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {API_KEY}"},
        json={
            "model": model,
            "max_tokens": 2048,
            "messages": messages,
            "tools": WORKER_TOOLS,
            "tool_choice": "auto",
        },
        timeout=API_TIMEOUT,
    )
    if r.status_code != 200:
        if pool:
            pool.fail(model)
        raise RuntimeError(f"HTTP {r.status_code}: {r.text[:300]}")
    data = r.json()
    if pool:
        pool.ok(model)
    return data


# ─── System prompt ──────────────────────────────────────────────────────────
GO_TEMPLATE = """package tools

import (
    "context"
    "fmt"
    "net/http"
    "io"
    "time"
)

// HandleXxx — replace Xxx with PascalCase name.
// Replaces: the original MCP server dependency.
// Tools exposed: list the tool names this handler serves.
func HandleXxx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    param, _ := getString(args, "param")
    if param == "" {
        return err("param is required")
    }
    client := &http.Client{Timeout: 30 * time.Second}
    req, e := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/?q="+param, nil)
    if e != nil {
        return err(fmt.Sprintf("request error: %v", e))
    }
    resp, e := client.Do(req)
    if e != nil {
        return err(fmt.Sprintf("fetch error: %v", e))
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    return ok(string(body))
}
"""

SYSTEM_PROMPT = (
    "You are an autonomous coding agent implementing Go tool modules for TormentNexus.\n"
    f"WORKSPACE: {WORKSPACE}\n"
    f"GO_TOOLS_DIR: {GO_DIR}\n"
    f"REGISTRY: {REGISTRY}\n\n"
"CRITICAL RULES — violation = immediate failure:\n"
"1. Your FIRST action MUST be a tool call (write_file, read_file, or bash).\n"
"2. Do NOT research. All specs are provided in the task. Implement immediately.\n"
"3. Write the Go file ONLY. Do NOT edit registry.go. Registration is handled automatically.\n"
"4. After creating any .go file, write a manifest JSON file at the SAME path but .json extension\n"
"   Example: if you create foo.go, write foo.go.json with: {\"tool_name\": \"handler_func\"} pairs\n"
"5. Build check is done automatically. If it fails, you will get the error — fix it.\n"
"6. Call done() when finished.\n\n"
"GO TOOL TEMPLATE:\n```go\n" + GO_TEMPLATE + "```\n\n"
"Available helpers (in parity.go — do NOT redefine):\n"
"- getString(args, keys...) → (string, bool)\n"
"- getInt(args, keys...) → int\n"
"- getBool(args, key) → bool\n"
"- ok(text) → (ToolResponse, error)\n"
"- err(text) → (ToolResponse, error)\n\n"
"MANIFEST FORMAT — for each Go file you create, also write a .json manifest:\n"
"   If you create tool_foo.go with HandleFoo and HandleBar, write tool_foo.go.json:\n"
'   {"tool_one": "HandleFoo", "tool_two": "HandleBar"}\n'
)


# ─── Pre-research: fetch READMEs and tool specs ─────────────────────────────
def research_servers(names_urls: list) -> dict:
    """Fetch README + tool names for a list of (name, github_url) pairs."""
    research = {}
    for name, url in names_urls:
        entry = {"readme": "", "tools": [], "api_endpoints": []}
        match = re.match(r"https://github\.com/([^/]+)/([^/\s#]+)", url)
        if match:
            owner, repo = match.group(1), match.group(2)
            try:
                resp = requests.get(
                    f"https://api.github.com/repos/{owner}/{repo}/readme",
                    headers={"Accept": "application/vnd.github.v3+json"},
                    timeout=15,
                )
                if resp.status_code == 200:
                    import base64
                    content = base64.b64decode(resp.json().get("content", "")).decode("utf-8", "replace")
                    entry["readme"] = content[:4000]
                    # Extract tool/function names
                    entry["tools"] = list(set(re.findall(
                        r'(?:tool|function|handler|action|capability|command)["\s:=]+["\']?(\w[\w_-]+)',
                        content, re.I
                    )))[:30]
                    # Extract API endpoints
                    entry["api_endpoints"] = re.findall(r'https?://[^\s"\'`\)]+', content)[:10]
            except Exception:
                pass
            time.sleep(0.5)  # GitHub rate limit
        research[name] = entry
    return research


# ─── Task generation ────────────────────────────────────────────────────────
def generate_tasks(limit: int, pre_research: bool = True) -> list:
    """Generate tasks from pending MCP servers in state DB."""
    conn = sqlite3.connect(STATE_DB)
    rows = conn.execute(
        "SELECT name, github_url, score FROM mcp_servers "
        "WHERE status='pending' ORDER BY score DESC LIMIT ?",
        (limit,),
    ).fetchall()
    conn.close()

    if not rows:
        print("No pending MCP servers found!")
        return []

    # Pre-research all servers at once
    research = {}
    if pre_research:
        print(f"Pre-researching {len(rows)} servers...")
        research = research_servers([(n, u) for n, u, _ in rows])
        researched = sum(1 for v in research.values() if v["readme"])
        print(f"Got READMEs for {researched}/{len(rows)} servers")

    tasks = []
    for name, url, score in rows:
        r = research.get(name, {})
        snake = re.sub(r"[^a-z0-9]+", "_", name.lower()).strip("_")
        readme = r.get("readme", "")
        tools = r.get("tools", [])
        endpoints = r.get("api_endpoints", [])

        # Build a rich task description with embedded research
        desc = f"Implement native Go module for MCP server: {name}\n\n"
        desc += f"Source: {url}\n"
        desc += f"Score: {score}/100\n"
        desc += f"Target file: {GO_DIR}/{snake}.go\n\n"

        if tools:
            desc += f"Tools this server exposes: {', '.join(tools[:20])}\n\n"
        if endpoints:
            desc += f"API endpoints found: {endpoints[:5]}\n\n"
        if readme:
            desc += f"README excerpt (use this to understand what to implement):\n---\n{readme[:3000]}\n---\n\n"

        pascal = snake.split('_')[0].title() + ''.join(w.title() for w in snake.split('_')[1:])
        desc += (
            f"INSTRUCTIONS:\n"
            f"1. Create {GO_DIR}/{snake}.go following the GO TOOL TEMPLATE\n"
            f"2. Name the main handler Handle{pascal}\n"
            f"3. Write a manifest file {GO_DIR}/{snake}.go.json with tool_name->handler mappings\n"
            f"4. Call done() with summary\n"
        )

        tasks.append({
            "id": f"mcp-{name}",
            "description": desc,
            "mcp_name": name,
            "github_url": url,
            "score": score,
            "_retries": 0,
            "_errors": [],
        })

    return tasks


# ─── Worker execution ───────────────────────────────────────────────────────
def run_task(task: dict, pool: ModelPool, wid: str) -> dict:
    """Run a single task with multi-turn LLM conversation. Returns result dict."""
    tid = task["id"]
    tag = f"[{wid}|{tid[:30]}]"
    model = pool.pick()
    print(f"{tag} START model={model}")

    # Build messages
    messages = [
        {"role": "system", "content": SYSTEM_PROMPT},
        {"role": "user", "content": task["description"]},
    ]

    # If this is a retry, add previous errors for context
    if task.get("_errors"):
        errors_str = "\n".join(task["_errors"][-3:])  # Last 3 errors
        messages.append({
            "role": "user",
            "content": f"PREVIOUS ATTEMPTS FAILED WITH THESE ERRORS — fix them:\n{errors_str}",
        })

    api_calls = 0
    t0 = time.time()
    files_touched = []
    summary = ""

    for turn in range(MAX_TURNS):
        # Rotate model every 5 turns
        if turn > 0 and turn % 5 == 0:
            model = pool.pick()

        # Call LLM with retry
        data = None
        for attempt in range(RETRY_LIMIT):
            try:
                data = llm_call(model, messages, pool)
                api_calls += 1
                break
            except Exception as e:
                err_str = str(e)[:200]
                print(f"{tag} API err attempt={attempt+1}: {err_str}")
                pool.fail(model)
                model = pool.pick()
                if attempt < RETRY_LIMIT - 1:
                    time.sleep(min(2 ** attempt, 16))

        if data is None:
            return {"id": tid, "status": "api_failed", "turns": turn, "elapsed": time.time() - t0}

        # Parse response
        choice = data.get("choices", [{}])[0]
        msg = choice.get("message", {})
        tool_calls = msg.get("tool_calls", [])
        content = msg.get("content", "") or ""

        # Add assistant message
        assistant = {"role": "assistant", "content": content}
        if tool_calls:
            assistant["tool_calls"] = tool_calls
        messages.append(assistant)

        # No tool calls
        if not tool_calls:
            if turn == 0:
                # Force the worker to use tools
                messages.append({
                    "role": "user",
                    "content": "You MUST call a tool NOW. Start with write_file to create the Go file, or bash to check existing code.",
                })
                continue
            # Natural end
            summary = content[:500]
            break

        # Process tool calls
        task_done = False
        for tc in tool_calls:
            fn = tc.get("function", {})
            tname = fn.get("name", "")
            try:
                targs = json.loads(fn.get("arguments", "{}"))
            except json.JSONDecodeError:
                targs = {}

            result = exec_tool(tname, targs)

            if result == "__DONE__":
                task_done = True
                summary = targs.get("summary", "Completed")
                files_touched = targs.get("files", files_touched)
                result = "Task marked complete."

            elif tname == "write_file":
                fp = targs.get("path", "")
                if fp.endswith(".go"):
                    files_touched.append(fp)

            messages.append({"role": "tool", "tool_call_id": tc.get("id", ""), "content": result})

        # Log progress
        names = [tc.get("function", {}).get("name", "?") for tc in tool_calls]
        elapsed = time.time() - t0
        print(f"{tag} T{turn+1}: {names} ({elapsed:.0f}s)")

        if task_done:
            return {
                "id": tid, "status": "completed", "summary": summary,
                "files": files_touched, "api_calls": api_calls,
                "turns": turn + 1, "elapsed": time.time() - t0, "model": model,
            }

    # Max turns or natural end
    status = "completed" if summary else "max_turns"
    return {
        "id": tid, "status": status, "summary": summary or "Max turns reached",
        "files": files_touched, "api_calls": api_calls,
        "turns": MAX_TURNS, "elapsed": time.time() - t0, "model": model,
    }


# ─── Build verification ─────────────────────────────────────────────────────
_registry_lock = threading.Lock()


def verify_build() -> tuple:
    """Run go build and return (success, error_text)."""
    r = subprocess.run(
        BUILD_CMD, shell=True, capture_output=True, text=True,
        timeout=120, encoding="utf-8", errors="replace",
    )
    ok = r.returncode == 0
    err = r.stderr if r.stderr else ""
    return ok, err


def merge_manifests_into_registry():
    """Read all .go.json manifest files and add handler registrations to registry.go."""
    with _registry_lock:
        manifests = sorted(Path(GO_DIR).glob("*.go.json"))
        if not manifests:
            return

        # Read current registry
        reg = Path(REGISTRY)
        content = reg.read_text(encoding="utf-8", errors="replace")

        # Find existing registrations (to avoid duplicates)
        existing = set(re.findall(r'r\.handlers\["([^"]+)"\]', content))

        # Collect all new registrations from manifests
        new_lines = []
        for mf in manifests:
            try:
                data = json.loads(mf.read_text(encoding="utf-8", errors="replace"))
                for tool_name, handler_func in data.items():
                    if tool_name not in existing:
                        new_lines.append(f'\tr.handlers["{tool_name}"] = {handler_func}')
                        existing.add(tool_name)
                # Remove manifest after processing
                mf.unlink()
            except Exception as e:
                print(f"  Manifest error {mf}: {e}")

        if not new_lines:
            return

        # Insert new registrations before the closing brace of registerAll()
        # Find the line with just "}" that closes registerAll
        lines = content.split("\n")
        # Find the closing brace of registerAll — it's after the last r.handlers line
        insert_idx = None
        for i in range(len(lines) - 1, -1, -1):
            if 'r.handlers[' in lines[i]:
                insert_idx = i + 1
                break

        if insert_idx is None:
            # Fallback: find the second closing brace (registerAll's)
            brace_count = 0
            for i, line in enumerate(lines):
                if 'func (r *Registry) registerAll()' in line:
                    brace_count = 1
                    continue
                if brace_count == 1:
                    if line.strip() == '}':
                        insert_idx = i
                        break

        if insert_idx:
            # Add comment header if this is the first batch
            header = "\n\t// Auto-registered by swarm\n"
            lines.insert(insert_idx, header + "\n".join(new_lines))
            reg.write_text("\n".join(lines), encoding="utf-8")
            print(f"  Merged {len(new_lines)} handler registrations into registry.go")


def auto_fix_build(error_text: str):
    """On build failure, try to identify and remove the problematic file."""
    # Extract the failing file from the error
    match = re.search(r'([^\s]+\.go):\d+', error_text)
    if match:
        bad_file = match.group(1)
        print(f"  Auto-fix: removing problematic file {bad_file}")
        p = Path(bad_file)
        if p.exists():
            p.unlink()
            # Also remove its registrations from registry.go
            reg = Path(REGISTRY)
            content = reg.read_text(encoding="utf-8", errors="replace")
            # Remove lines referencing the deleted file's handler
            base = p.stem.replace(".go", "")
            lines = content.split("\n")
            filtered = [l for l in lines if base.lower() not in l.lower() or 'r.handlers[' not in l]
            reg.write_text("\n".join(filtered), encoding="utf-8")
            # Re-check build
            ok, err = verify_build()
            if ok:
                print(f"  Auto-fix: build recovered after removing {bad_file}")
            else:
                print(f"  Auto-fix: build still failing: {err[:200]}")


# ─── State DB updates ───────────────────────────────────────────────────────
def mark_server(name, status, go_file=""):
    conn = sqlite3.connect(STATE_DB)
    conn.execute(
        "UPDATE mcp_servers SET status=?, go_file=?, updated_at=CURRENT_TIMESTAMP WHERE name=?",
        (status, go_file, name),
    )
    conn.commit()
    conn.close()


# ─── Main orchestrator ──────────────────────────────────────────────────────
def main():
    parser = argparse.ArgumentParser(description="TormentNexus Swarm v2")
    parser.add_argument("--workers", type=int, default=5)
    parser.add_argument("--limit", type=int, default=50)
    parser.add_argument("--model", type=str, default=None)
    parser.add_argument("--no-research", action="store_true", help="Skip pre-research phase")
    parser.add_argument("--forever", action="store_true", help="Keep running, re-polling for new tasks")
    args = parser.parse_args()

    if args.model:
        global MODELS
        MODELS = [args.model]

    # Reduce workers based on available models
    actual_workers = min(args.workers, len(MODELS) + 1)
    if actual_workers < args.workers:
        print(f'  Reducing workers from {args.workers} to {actual_workers} (limited by {len(MODELS)} models)')
        args.workers = actual_workers

    pool = ModelPool(MODELS)

    while True:  # Forever loop
        # Health check: verify proxy can handle requests
        print('Checking proxy health...')
        try:
            r = requests.post(
                f'{PROXY}/v1/chat/completions',
                headers={'Content-Type': 'application/json', 'Authorization': f'Bearer {API_KEY}'},
                json={'model': MODELS[0], 'max_tokens': 32, 'messages': [{'role': 'user', 'content': 'Say OK'}]},
                timeout=120,
            )
            if r.status_code == 200:
                print(f'  Proxy healthy: {MODELS[0]} responded')
            else:
                print(f'  Proxy returned HTTP {r.status_code}, may have issues')
        except Exception as e:
            print(f'  Proxy not responding: {e}')
            print('  Waiting 60s for recovery...')
            time.sleep(60)

        # Generate tasks
        tasks = generate_tasks(args.limit, pre_research=not args.no_research)
        if not tasks:
            if not args.forever:
                print("No tasks. Exiting.")
                break
            print("No tasks. Sleeping 60s...")
            time.sleep(60)
            continue

        # Stats
        stats = {"done": 0, "failed": 0, "active": 0, "t0": time.time()}
        results_all = []
        lock = threading.Lock()

        # Stats reporter
        stop_evt = threading.Event()
        def reporter():
            while not stop_evt.wait(20):
                with lock:
                    el = time.time() - stats["t0"]
                    rate = stats["done"] / max(el / 60, 0.01)
                    print(f"[STATS] {el:.0f}s done={stats['done']} fail={stats['failed']} "
                          f"active={stats['active']} pending={len(tasks)} rate={rate:.1f}/min")
        threading.Thread(target=reporter, daemon=True).start()

        print(f"\n{'='*60}")
        print(f"  TORMENTNEXUS SWARM v2")
        print(f"  Workers: {args.workers}  Models: {len(MODELS)}  Tasks: {len(tasks)}")
        print(f"  Models: {MODELS}")
        print(f"{'='*60}\n")

        # Worker function
        def worker(wid):
            while True:
                with lock:
                    if not tasks:
                        break
                    task = tasks.pop(0)
                    stats["active"] += 1

                result = run_task(task, pool, wid)

                with lock:
                    stats["active"] -= 1
                    results_all.append(result)

                # Handle retry
                if result["status"] not in ("completed",):
                    task["_retries"] += 1
                    # Capture error context for retry
                    if result.get("summary"):
                        task["_errors"].append(f"Attempt failed ({result['status']}): {result['summary'][:300]}")

                    if task["_retries"] < RETRY_LIMIT:
                        with lock:
                            tasks.append(task)  # Re-queue
                        print(f"[{wid}] {task['id'][:30]} failed, retry #{task['_retries']}")
                    else:
                        with lock:
                            stats["failed"] += 1
                        # Mark in DB
                        mname = task.get("mcp_name", "")
                        if mname:
                            mark_server(mname, "failed")
                        print(f"[{wid}] {task['id'][:30]} FAILED after {task['_retries']} retries")
                else:
                    with lock:
                        stats["done"] += 1
                    mname = task.get("mcp_name", "")
                    go_file = ""
                    for f in result.get("files", []):
                        if f.endswith(".go") and "registry" not in f.lower():
                            go_file = f
                            break
                    if mname:
                        mark_server(mname, "implemented", go_file)
                    print(f"[{wid}] {task['id'][:30]} DONE "
                          f"({result.get('turns',0)}t, {result.get('elapsed',0):.0f}s)")

        # Merge manifests into registry.go and build-check
        merge_manifests_into_registry()
        build_ok, build_err = verify_build()
        if not build_ok:
            print(f"  Post-merge build failed: {build_err[:200]}")
            # Try to fix by removing the last merged file
            auto_fix_build(build_err)

        # Launch workers
        with ThreadPoolExecutor(max_workers=args.workers) as ex:
            futures = [ex.submit(worker, f"W{i+1}") for i in range(args.workers)]
            for f in as_completed(futures):
                try:
                    f.result()
                except Exception as e:
                    print(f"Worker crashed: {e}")
                    traceback.print_exc()

        stop_evt.set()

        # Merge any remaining manifests
        merge_manifests_into_registry()

        # Final report
        elapsed = time.time() - stats["t0"]
        completed = [r for r in results_all if r["status"] == "completed"]
        failed = [r for r in results_all if r["status"] != "completed"]

        print(f"\n{'='*60}")
        print(f"  SWARM COMPLETE")
        print(f"  Completed: {len(completed)}/{len(results_all)}")
        print(f"  Failed: {len(failed)}")
        print(f"  Time: {elapsed:.0f}s ({elapsed/60:.1f}min)")
        print(f"  Rate: {len(completed)/max(elapsed/60,0.01):.1f}/min")
        print(f"{'='*60}")

        # Verify Go build integrity
        print("\nFinal build verification...")
        build_ok, build_err = verify_build()
        if build_ok:
            print("  BUILD CLEAN")
        else:
            print(f"  BUILD FAILED:\n{build_err[:500]}")

        # Save results
        results_file = Path(WORKSPACE) / "swarm_results.json"
        results_file.write_text(json.dumps({
            "completed": len(completed),
            "failed": len(failed),
            "build_ok": build_ok,
            "results": results_all,
        }, indent=2))
        print(f"Results saved to {results_file}")

        if not args.forever:
            break

        print("\nForever mode: sleeping 30s then polling for new tasks...")
        time.sleep(30)


if __name__ == "__main__":
    main()
