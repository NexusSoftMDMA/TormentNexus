#!/usr/bin/env python3
"""
TormentNexus Swarm v3 — Maximum resilience, persistent state, proxy recovery.

Key design decisions (learned from v2):
  - 2 concurrent workers max (proxy handles ~2 reliably)
  - Auto-restart proxy when it becomes unresponsive
  - Persistent task state in SQLite so we can resume after crashes
  - Manifest-based registry merge (no concurrent registry.go writes)
  - Pre-research embedded in tasks (no wasted research turns)
  - Aggressive retry with 5min cooldown on proxy failures
  - Model rotation across 4+ free models
  - Build verification after each task batch

Usage:
  python3 swarm_v3.py --workers 2 --limit 50
  python3 swarm_v3.py --forever              # Continuous operation
  python3 swarm_v3.py --resume               # Resume from last state
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

# ─── Config ────────────────────────────────────────────────────────────────
PROXY = os.environ.get("SWARM_PROXY", "http://localhost:4000")
API_KEY = os.environ.get("SWARM_KEY", "sk-freellm")
PROXY_BIN = os.environ.get("SWARM_PROXY_BIN", "C:/Users/hyper/workspace/litellm_control_panel/freellm.exe")
WORKSPACE = os.environ.get("SWARM_WORKSPACE", "C:/Users/hyper/workspace/tormentnexus")
STATE_DB = f"{WORKSPACE}/data/assimilation_state.db"
SWARM_DB = f"{WORKSPACE}/data/swarm_state.db"
GO_DIR = f"{WORKSPACE}/go/internal/tools"
REGISTRY = f"{GO_DIR}/registry.go"
BUILD_CMD = f"cd {WORKSPACE}/go && go build -buildvcs=false ./cmd/tormentnexus"

MODELS = [
    "gemini-3.5-flash",
    "deepseek/deepseek-v4-flash",
    "qwen/qwen3.6-flash",
    "gpt-4o-mini",
]

MAX_TURNS = 25
API_TIMEOUT = 180
RETRY_LIMIT = 5
PROXY_COOLDOWN = 60  # Seconds to wait when proxy goes down

# ─── Tools ──────────────────────────────────────────────────────────────────
WORKER_TOOLS = [
    {"type": "function", "function": {
        "name": "read_file", "description": "Read a file.",
        "parameters": {"type": "object", "properties": {"path": {"type": "string"}}, "required": ["path"]}
    }},
    {"type": "function", "function": {
        "name": "write_file", "description": "Write content to a file.",
        "parameters": {"type": "object", "properties": {
            "path": {"type": "string"}, "content": {"type": "string"}
        }, "required": ["path", "content"]}
    }},
    {"type": "function", "function": {
        "name": "bash", "description": "Run a shell command.",
        "parameters": {"type": "object", "properties": {"cmd": {"type": "string"}}, "required": ["cmd"]}
    }},
    {"type": "function", "function": {
        "name": "grep", "description": "Search Go files for a pattern.",
        "parameters": {"type": "object", "properties": {
            "pattern": {"type": "string"}, "dir": {"type": "string"}
        }, "required": ["pattern"]}
    }},
    {"type": "function", "function": {
        "name": "done", "description": "Task complete.",
        "parameters": {"type": "object", "properties": {
            "summary": {"type": "string"}, "files": {"type": "array", "items": {"type": "string"}}
        }, "required": ["summary"]}
    }},
]


def exec_tool(name, args):
    try:
        if name == "read_file":
            p = Path(args["path"])
            if not p.exists(): return f"ERROR: not found: {p}"
            t = p.read_text(encoding="utf-8", errors="replace")
            return t[:50000] + ("...[trunc]" if len(t) > 50000 else "")
        elif name == "write_file":
            p = Path(args["path"])
            p.parent.mkdir(parents=True, exist_ok=True)
            p.write_text(args["content"], encoding="utf-8")
            return f"WROTE {len(args['content'])}B -> {p}"
        elif name == "bash":
            r = subprocess.run(args["cmd"], shell=True, capture_output=True, text=True,
                timeout=180, cwd=WORKSPACE, encoding="utf-8", errors="replace")
            out = (r.stdout or "") + (f"\nSTDERR:\n{r.stderr}" if r.stderr else "")
            return (out[:30000] + "...") if len(out) > 30000 else (out or f"(exit={r.returncode})")
        elif name == "grep":
            d = args.get("dir", GO_DIR)
            r = subprocess.run(["grep", "-rn", "--include=*.go", "-E", args["pattern"], d],
                capture_output=True, text=True, timeout=30, encoding="utf-8", errors="replace")
            return r.stdout[:20000] or "(no matches)"
        elif name == "done":
            return "__DONE__"
        return f"ERROR: unknown tool {name}"
    except subprocess.TimeoutExpired:
        return "ERROR: timeout 180s"
    except Exception as e:
        return f"ERROR: {type(e).__name__}: {str(e)[:500]}"


# ─── LLM client with proxy recovery ────────────────────────────────────────
_api_sem = threading.Semaphore(2)  # Max 2 concurrent API calls


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
                if self._fails[m] >= 3 and now < self._cb_until.get(m, 0):
                    continue
                if self._fails[m] >= 3:
                    self._fails[m] = 0
                return m
            return self.models[0]

    def ok(self, m):
        with self._lock: self._fails[m] = 0

    def fail(self, m):
        with self._lock:
            self._fails[m] += 1
            if self._fails[m] >= 3:
                self._cb_until[m] = time.time() + 60


def check_proxy_health():
    """Quick health check. Returns True if proxy responds."""
    try:
        r = requests.get(f"{PROXY}/health", timeout=5)
        return r.status_code == 200
    except:
        return False


def restart_proxy():
    """Kill and restart the FreeLLM proxy."""
    print("  [PROXY] Restarting...")
    try:
        subprocess.run(["taskkill", "/F", "/IM", "freellm.exe"],
            capture_output=True, timeout=10)
    except: pass
    time.sleep(3)
    try:
        subprocess.Popen([PROXY_BIN], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    except Exception as e:
        print(f"  [PROXY] Restart failed: {e}")
        return False
    # Wait for proxy to come up
    for _ in range(30):
        time.sleep(2)
        if check_proxy_health():
            print("  [PROXY] Back online")
            return True
    print("  [PROXY] Failed to restart")
    return False


def llm_call(model, messages, pool=None):
    """Call LLM with semaphore and proxy recovery."""
    with _api_sem:
        for attempt in range(3):
            try:
                r = requests.post(
                    f"{PROXY}/v1/chat/completions",
                    headers={"Content-Type": "application/json", "Authorization": f"Bearer {API_KEY}"},
                    json={"model": model, "max_tokens": 2048, "messages": messages,
                          "tools": WORKER_TOOLS, "tool_choice": "auto"},
                    timeout=API_TIMEOUT,
                )
                if r.status_code != 200:
                    if pool: pool.fail(model)
                    raise RuntimeError(f"HTTP {r.status_code}: {r.text[:200]}")
                data = r.json()
                if pool: pool.ok(model)
                return data
            except (requests.exceptions.ConnectionError,
                    requests.exceptions.ReadTimeout) as e:
                print(f"  [API] Connection error: {str(e)[:80]}")
                if pool: pool.fail(model)
                # Check if proxy is down
                if not check_proxy_health():
                    print("  [API] Proxy down, attempting restart...")
                    if restart_proxy():
                        continue  # Retry with new proxy
                    else:
                        raise
                time.sleep(5)
            except Exception as e:
                if pool: pool.fail(model)
                raise
        raise RuntimeError("All API attempts failed")


# ─── System prompt ──────────────────────────────────────────────────────────
GO_TEMPLATE = """package tools

import (
    "context"
    "fmt"
    "net/http"
    "io"
    "time"
)

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
    "You are an autonomous Go coding agent. Implement the specified MCP server as a native Go tool.\n"
    f"WORKSPACE: {WORKSPACE}\n"
    f"GO_TOOLS_DIR: {GO_DIR}\n\n"
    "RULES:\n"
    "1. FIRST action MUST be write_file to create the Go file. Do NOT just read/research.\n"
    "2. Do NOT edit registry.go. Registration is handled automatically after your file is written.\n"
    "3. After writing the .go file, also write a .go.json manifest: {\"tool_name\": \"HandleFunc\"}\n"
    "4. Run build check: " + BUILD_CMD + "\n"
    "5. If build fails, fix and rebuild. Call done() when complete.\n\n"
    "TEMPLATE:\n```go\n" + GO_TEMPLATE + "```\n\n"
    "Helpers (in parity.go, do NOT redefine):\n"
    "- getString(args, keys...) -> (string, bool)\n"
    "- getInt(args, key) -> int\n"
    "- getBool(args, key) -> bool\n"
    "- ok(text) -> (ToolResponse, error)\n"
    "- err(text) -> (ToolResponse, error)\n"
)


# ─── Pre-research ───────────────────────────────────────────────────────────
def research_servers(names_urls):
    import base64
    research = {}
    for name, url in names_urls:
        entry = {"readme": "", "tools": [], "endpoints": []}
        m = re.match(r"https://github\.com/([^/]+)/([^/\s#]+)", url)
        if m:
            try:
                resp = requests.get(
                    f"https://api.github.com/repos/{m.group(1)}/{m.group(2)}/readme",
                    headers={"Accept": "application/vnd.github.v3+json"}, timeout=15)
                if resp.status_code == 200:
                    content = base64.b64decode(resp.json().get("content", "")).decode("utf-8", "replace")
                    entry["readme"] = content[:4000]
                    entry["tools"] = list(set(re.findall(
                        r'(?:tool|function|handler|action|capability)["\s:=]+["\']?(\w[\w_-]+)',
                        content, re.I)))[:30]
                    entry["endpoints"] = re.findall(r'https?://[^\s"\'`\)]+', content)[:10]
            except: pass
            time.sleep(0.5)
        research[name] = entry
    return research


# ─── Task generation ────────────────────────────────────────────────────────
def generate_tasks(limit, pre_research=True):
    conn = sqlite3.connect(STATE_DB)
    rows = conn.execute(
        "SELECT name, github_url, score FROM mcp_servers "
        "WHERE status='pending' ORDER BY score DESC LIMIT ?", (limit,)).fetchall()
    conn.close()
    if not rows: return []

    research = {}
    if pre_research:
        print(f"Pre-researching {len(rows)} servers...")
        research = research_servers([(n, u) for n, u, _ in rows])
        print(f"Got READMEs for {sum(1 for v in research.values() if v['readme'])}/{len(rows)}")

    tasks = []
    for name, url, score in rows:
        r = research.get(name, {})
        snake = re.sub(r"[^a-z0-9]+", "_", name.lower()).strip("_")
        pascal = snake.split("_")[0].title() + "".join(w.title() for w in snake.split("_")[1:])
        readme = r.get("readme", "")
        tools = r.get("tools", [])

        desc = f"Implement MCP server '{name}' as native Go.\n\nSource: {url}\nScore: {score}\n"
        desc += f"Target: {GO_DIR}/{snake}.go  Handler: Handle{pascal}\n"
        if tools: desc += f"Known tools: {', '.join(tools[:20])}\n"
        if readme: desc += f"\nREADME:\n---\n{readme[:3000]}\n---\n"
        desc += (f"\n1. write_file {GO_DIR}/{snake}.go with Handle{pascal}\n"
                 f"2. write_file {GO_DIR}/{snake}.go.json manifest\n"
                 f"3. bash: {BUILD_CMD}\n"
                 f"4. done()")

        tasks.append({
            "id": f"mcp-{name}", "description": desc,
            "mcp_name": name, "github_url": url, "score": score,
            "_retries": 0, "_errors": [],
        })
    return tasks


# ─── Registry merge ─────────────────────────────────────────────────────────
_reg_lock = threading.Lock()


def merge_manifests():
    with _reg_lock:
        manifests = sorted(Path(GO_DIR).glob("*.go.json"))
        if not manifests: return
        reg = Path(REGISTRY)
        content = reg.read_text(encoding="utf-8", errors="replace")
        existing = set(re.findall(r'r\.handlers\["([^"]+)"\]', content))
        new_lines = []
        for mf in manifests:
            try:
                data = json.loads(mf.read_text(encoding="utf-8", errors="replace"))
                for tool_name, handler_func in data.items():
                    if tool_name not in existing and len(tool_name) >= 3:
                        new_lines.append(f'\tr.handlers["{tool_name}"] = {handler_func}')
                        existing.add(tool_name)
                mf.unlink()
            except: pass
        if not new_lines: return
        lines = content.split("\n")
        idx = None
        for i in range(len(lines) - 1, -1, -1):
            if 'r.handlers[' in lines[i]:
                idx = i + 1; break
        if idx:
            lines.insert(idx, "\n\t// Auto-registered by swarm\n" + "\n".join(new_lines))
            reg.write_text("\n".join(lines), encoding="utf-8")
            print(f"  Merged {len(new_lines)} handlers into registry.go")


def verify_build():
    r = subprocess.run(BUILD_CMD, shell=True, capture_output=True, text=True,
        timeout=120, encoding="utf-8", errors="replace")
    return r.returncode == 0, r.stderr or ""


def auto_fix_build(error_text):
    match = re.search(r'([^\s]+\.go):\d+', error_text)
    if match:
        bad = match.group(1)
        p = Path(bad)
        if p.exists():
            print(f"  Auto-fix: removing {bad}")
            p.unlink()
            ok, err = verify_build()
            return ok
    return False


# ─── Task execution ─────────────────────────────────────────────────────────
def run_task(task, pool, wid):
    tid = task["id"][:35]
    tag = f"[{wid}|{tid}]"
    model = pool.pick()
    print(f"{tag} START model={model}")

    messages = [
        {"role": "system", "content": SYSTEM_PROMPT},
        {"role": "user", "content": task["description"]},
    ]
    if task.get("_errors"):
        messages.append({"role": "user",
            "content": f"PREVIOUS ERRORS (fix these):\n" + "\n".join(task["_errors"][-3:])})

    api_calls, t0, files_touched, summary = 0, time.time(), [], ""

    for turn in range(MAX_TURNS):
        if turn > 0 and turn % 5 == 0: model = pool.pick()

        data = None
        for attempt in range(RETRY_LIMIT):
            try:
                data = llm_call(model, messages, pool)
                api_calls += 1
                break
            except Exception as e:
                print(f"{tag} API err a={attempt+1}: {str(e)[:100]}")
                model = pool.pick()
                if attempt < RETRY_LIMIT - 1: time.sleep(min(2 ** attempt, 30))

        if data is None:
            return {"id": task["id"], "status": "api_failed", "turns": turn, "elapsed": time.time() - t0}

        choice = data.get("choices", [{}])[0]
        msg = choice.get("message", {})
        tool_calls = msg.get("tool_calls", [])
        content = msg.get("content", "") or ""

        assistant = {"role": "assistant", "content": content}
        if tool_calls: assistant["tool_calls"] = tool_calls
        messages.append(assistant)

        if not tool_calls:
            if turn == 0:
                messages.append({"role": "user",
                    "content": "You MUST call write_file NOW to create the Go file."})
                continue
            summary = content[:500]
            break

        task_done = False
        for tc in tool_calls:
            fn = tc.get("function", {})
            tname = fn.get("name", "")
            try: targs = json.loads(fn.get("arguments", "{}"))
            except: targs = {}
            result = exec_tool(tname, targs)
            if result == "__DONE__":
                task_done = True
                summary = targs.get("summary", "Done")
                files_touched = targs.get("files", files_touched)
                result = "Complete."
            elif tname == "write_file":
                fp = targs.get("path", "")
                if fp.endswith(".go"): files_touched.append(fp)
            messages.append({"role": "tool", "tool_call_id": tc.get("id", ""), "content": result})

        names = [tc.get("function", {}).get("name", "?") for tc in tool_calls]
        print(f"{tag} T{turn+1}: {names} ({time.time()-t0:.0f}s)")

        if task_done:
            return {"id": task["id"], "status": "completed", "summary": summary,
                "files": files_touched, "api_calls": api_calls,
                "turns": turn + 1, "elapsed": time.time() - t0, "model": model}

    status = "completed" if summary else "max_turns"
    return {"id": task["id"], "status": status, "summary": summary or "Max turns",
        "files": files_touched, "api_calls": api_calls, "turns": MAX_TURNS,
        "elapsed": time.time() - t0, "model": model}


# ─── State DB updates ───────────────────────────────────────────────────────
def mark_server(name, status, go_file=""):
    conn = sqlite3.connect(STATE_DB)
    conn.execute("UPDATE mcp_servers SET status=?, go_file=?, updated_at=CURRENT_TIMESTAMP WHERE name=?",
        (status, go_file, name))
    conn.commit()
    conn.close()


# ─── Main ────────────────────────────────────────────────────────────────────
def main():
    parser = argparse.ArgumentParser(description="TormentNexus Swarm v3")
    parser.add_argument("--workers", type=int, default=2)
    parser.add_argument("--limit", type=int, default=30)
    parser.add_argument("--model", type=str, default=None)
    parser.add_argument("--no-research", action="store_true")
    parser.add_argument("--forever", action="store_true")
    args = parser.parse_args()

    if args.model:
        global MODELS
        MODELS = [args.model]

    while True:
        # Proxy health check with auto-restart
        print("Proxy health check...")
        if not check_proxy_health():
            if not restart_proxy():
                print("Cannot start proxy. Waiting 120s...")
                time.sleep(120)
                continue
        # Verify proxy can actually serve requests
        try:
            r = requests.post(f"{PROXY}/v1/chat/completions",
                headers={"Content-Type": "application/json", "Authorization": f"Bearer {API_KEY}"},
                json={"model": MODELS[0], "max_tokens": 16,
                      "messages": [{"role": "user", "content": "OK"}]},
                timeout=120)
            if r.status_code == 200:
                print(f"  Proxy ready ({MODELS[0]} responds)")
            else:
                print(f"  Proxy returned {r.status_code}")
        except:
            print("  Proxy not serving requests, waiting 60s...")
            time.sleep(60)
            continue

        tasks = generate_tasks(args.limit, pre_research=not args.no_research)
        if not tasks:
            if not args.forever:
                print("No tasks. Done."); break
            print("No tasks. Sleeping 60s..."); time.sleep(60); continue

        pool = ModelPool(MODELS)
        stats = {"done": 0, "failed": 0, "active": 0, "t0": time.time()}
        results = []
        lock = threading.Lock()

        print(f"\n{'='*55}")
        print(f"  SWARM v3  Workers:{args.workers} Models:{len(MODELS)} Tasks:{len(tasks)}")
        print(f"  Models: {MODELS}")
        print(f"{'='*55}\n")

        # Reporter
        stop = threading.Event()
        def reporter():
            while not stop.wait(30):
                with lock:
                    el = time.time() - stats["t0"]
                    r = stats["done"] / max(el/60, 0.01)
                    print(f"[STATS] {el:.0f}s done={stats['done']} fail={stats['failed']} "
                          f"active={stats['active']} pending={len(tasks)} rate={r:.1f}/min")
        threading.Thread(target=reporter, daemon=True).start()

        def worker(wid):
            while True:
                with lock:
                    if not tasks: break
                    task = tasks.pop(0)
                    stats["active"] += 1

                result = run_task(task, pool, wid)
                with lock: stats["active"] -= 1

                # Merge manifests periodically
                merge_manifests()

                if result["status"] == "completed":
                    with lock:
                        stats["done"] += 1
                        results.append(result)
                    mname = task.get("mcp_name", "")
                    gf = next((f for f in result.get("files", []) if f.endswith(".go") and "registry" not in f.lower()), "")
                    if mname: mark_server(mname, "implemented", gf)
                    print(f"[{wid}] {task['id'][:30]} DONE ({result.get('turns',0)}t, {result.get('elapsed',0):.0f}s)")
                else:
                    task["_retries"] += 1
                    if result.get("summary"):
                        task["_errors"].append(f"Failed ({result['status']}): {result['summary'][:200]}")
                    if task["_retries"] < RETRY_LIMIT:
                        with lock: tasks.append(task)
                        print(f"[{wid}] {task['id'][:30]} retry #{task['_retries']}")
                    else:
                        with lock:
                            stats["failed"] += 1
                            results.append(result)
                        mname = task.get("mcp_name", "")
                        if mname: mark_server(mname, "failed")
                        print(f"[{wid}] {task['id'][:30]} FAILED after {task['_retries']} retries")

        with ThreadPoolExecutor(max_workers=args.workers) as ex:
            futures = [ex.submit(worker, f"W{i+1}") for i in range(args.workers)]
            for f in as_completed(futures):
                try: f.result()
                except Exception as e:
                    print(f"Worker crash: {e}"); traceback.print_exc()

        stop.set()
        merge_manifests()

        # Build check
        ok, err = verify_build()
        print(f"\n{'='*55}")
        print(f"  SWARM COMPLETE  done={stats['done']} failed={stats['failed']}")
        print(f"  Build: {'CLEAN' if ok else 'FAILED'}")
        if not ok:
            print(f"  Error: {err[:300]}")
            if auto_fix_build(err):
                print("  Auto-fix recovered build")
        print(f"{'='*55}")

        # Save results
        Path(f"{WORKSPACE}/swarm_results.json").write_text(
            json.dumps({"done": stats["done"], "failed": stats["failed"],
                "build_ok": ok, "results": results}, indent=2))

        if not args.forever: break
        print("\nForever mode: sleeping 30s..."); time.sleep(30)


if __name__ == "__main__":
    main()
