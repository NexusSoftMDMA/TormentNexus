import subprocess
#!/usr/bin/env python3
"""
TormentNexus Swarm — Parallel LLM Workers via FreeLLM Proxy

Architecture:
  - Uses OpenAI /v1/chat/completions endpoint directly (no SDK)
  - Rotates across multiple free models to spread rate-limit load
  - Multi-turn tool use with automatic tool-result injection
  - Registry.go serialization via file lock (only 1 writer at a time)
  - Exponential backoff + per-model circuit breaker
  - Tasks read from SQLite state DB, results written back

Usage:
  python3 swarm.py --workers 5 --track mcp --limit 50
  python3 swarm.py --workers 3 --track hermes --limit 30
  python3 swarm.py --workers 2 --tasks tasks.json

Models are rotated automatically from SWARM_MODELS env or defaults.
"""

import argparse
import json
import os
import re
import sqlite3
import threading
import time
import traceback
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor, as_completed

import requests

# ─── Configuration ────────────────────────────────────────────────────────
PROXY_BASE = os.environ.get("SWARM_PROXY", "http://localhost:4000")
PROXY_KEY = os.environ.get("SWARM_PROXY_KEY", "sk-freellm")
WORKSPACE = os.environ.get("SWARM_WORKSPACE", "C:/Users/hyper/workspace/tormentnexus")
STATE_DB = os.environ.get("SWARM_STATE_DB", f"{WORKSPACE}/data/assimilation_state.db")
GO_TOOLS_DIR = f"{WORKSPACE}/go/internal/tools"
REGISTRY_FILE = f"{GO_TOOLS_DIR}/registry.go"

# Models that support multi-turn tool use through OpenAI format
DEFAULT_MODELS = [
    "free-llm",
    "free-llm-fallback",
    "deepseek/deepseek-v4-flash",
    "qwen/qwen3.6-flash",
    "gpt-4o-mini",
]

MODELS = (
    os.environ.get("SWARM_MODELS", "").split(",")
    if os.environ.get("SWARM_MODELS")
    else DEFAULT_MODELS
)

API_TIMEOUT = int(os.environ.get("SWARM_API_TIMEOUT", "120"))
MAX_TURNS = int(os.environ.get("SWARM_MAX_TURNS", "25"))
MAX_RETRIES = int(os.environ.get("SWARM_MAX_RETRIES", "5"))
CIRCUIT_BREAKER_THRESHOLD = int(os.environ.get("SWARM_CB_THRESHOLD", "3"))
CIRCUIT_BREAKER_RESET = int(os.environ.get("SWARM_CB_RESET", "120"))

# ─── Tool Definitions ─────────────────────────────────────────────────────
SWARM_TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "read_file",
            "description": "Read the contents of a file. Returns the full text content.",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {"type": "string", "description": "File path to read"}
                },
                "required": ["path"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "write_file",
            "description": "Write content to a file. Creates parent directories if needed.",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {"type": "string", "description": "File path to write"},
                    "content": {"type": "string", "description": "Content to write"},
                },
                "required": ["path", "content"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "run_bash",
            "description": "Execute a bash command and return stdout/stderr. Use for builds, tests, grep, etc.",
            "parameters": {
                "type": "object",
                "properties": {
                    "command": {"type": "string", "description": "Bash command to run"}
                },
                "required": ["command"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "grep_search",
            "description": "Search for a pattern in files. Returns matching lines with file paths.",
            "parameters": {
                "type": "object",
                "properties": {
                    "pattern": {
                        "type": "string",
                        "description": "Regex or text pattern",
                    },
                    "path": {
                        "type": "string",
                        "description": "Directory to search (default: workspace root)",
                    },
                },
                "required": ["pattern"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "list_files",
            "description": "List files in a directory.",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {"type": "string", "description": "Directory path"},
                    "recursive": {"type": "boolean", "description": "List recursively"},
                },
                "required": ["path"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "task_complete",
            "description": "Call this when your task is fully done. Provide a summary of what you did and list any files you created/modified.",
            "parameters": {
                "type": "object",
                "properties": {
                    "summary": {
                        "type": "string",
                        "description": "Summary of work completed",
                    },
                    "files_changed": {
                        "type": "array",
                        "items": {"type": "string"},
                        "description": "List of files created or modified",
                    },
                    "discoveries": {
                        "type": "string",
                        "description": "Any useful patterns or gotchas discovered",
                    },
                },
                "required": ["summary"],
            },
        },
    },
]


# ─── Tool Execution ────────────────────────────────────────────────────────
def execute_tool(name: str, args: dict) -> str:
    """Execute a tool call locally and return the result as text."""
    try:
        if name == "read_file":
            path = args.get("path", "")
            if not path:
                return "Error: path is required"
            p = Path(path)
            if not p.exists():
                return f"Error: file not found: {path}"
            content = p.read_text(encoding="utf-8", errors="replace")
            # Truncate huge files
            if len(content) > 50000:
                content = (
                    content[:50000] + f"\n... [truncated, {len(content)} bytes total]"
                )
            return content

        elif name == "write_file":
            path = args.get("path", "")
            content = args.get("content", "")
            if not path:
                return "Error: path is required"
            p = Path(path)
            p.parent.mkdir(parents=True, exist_ok=True)
            p.write_text(content, encoding="utf-8")
            return f"OK: wrote {len(content)} bytes to {path}"

        elif name == "run_bash":
            cmd = args.get("command", "")
            if not cmd:
                return "Error: command is required"
            import subprocess

            result = subprocess.run(
                cmd,
                shell=True,
                capture_output=True,
                text=True,
                timeout=120,
                cwd=WORKSPACE,
                encoding="utf-8",
                errors="replace",
            )
            output = ""
            if result.stdout:
                output += result.stdout
            if result.stderr:
                output += ("\nSTDERR:\n" + result.stderr) if output else result.stderr
            if not output:
                output = f"(exit code {result.returncode}, no output)"
            if len(output) > 30000:
                output = output[:30000] + "\n... [truncated]"
            return output

        elif name == "grep_search":
            import subprocess

            pattern = args.get("pattern", "")
            search_path = args.get("path", WORKSPACE)
            result = subprocess.run(
                ["grep", "-rn", "--include=*.go", "-E", pattern, search_path],
                capture_output=True,
                text=True,
                timeout=30,
            )
            output = result.stdout or "(no matches)"
            if len(output) > 20000:
                output = output[:20000] + "\n... [truncated]"
            return output

        elif name == "list_files":
            import subprocess

            path = args.get("path", WORKSPACE)
            recursive = args.get("recursive", False)
            cmd = ["find", path, "-type", "f", "-name", "*.go"]
            if not recursive:
                cmd += ["-maxdepth", "1"]
            result = subprocess.run(cmd, capture_output=True, text=True, timeout=15,
                                  encoding='utf-8', errors='replace')
            output = result.stdout or "(empty)"
            if len(output) > 10000:
                output = output[:10000] + "\n... [truncated]"
            return output

        elif name == "task_complete":
            return "__TASK_COMPLETE__"

        else:
            return f"Error: unknown tool {name}"

    except subprocess.TimeoutExpired:
        return "Error: command timed out (120s)"
    except Exception as e:
        return f"Error: {type(e).__name__}: {str(e)[:500]}"


# ─── LLM Client ────────────────────────────────────────────────────────────
class ModelRotator:
    """Rotate across models with per-model circuit breaker."""

    def __init__(self, models: list):
        self.models = models
        self.fail_counts = {m: 0 for m in models}
        self.circuit_open_until = {m: 0 for m in models}
        self.lock = threading.Lock()
        self._idx = 0

    def next_model(self) -> str:
        with self.lock:
            now = time.time()
            # Try each model starting from current index
            for _ in range(len(self.models)):
                m = self.models[self._idx % len(self.models)]
                self._idx += 1
                # Check circuit breaker
                if self.fail_counts[m] >= CIRCUIT_BREAKER_THRESHOLD:
                    if now < self.circuit_open_until.get(m, 0):
                        continue  # Circuit still open
                    else:
                        # Reset circuit
                        self.fail_counts[m] = 0
                        return m
                else:
                    return m
            # All circuits open — just use first model and hope
            return self.models[0]

    def report_success(self, model: str):
        with self.lock:
            self.fail_counts[model] = 0

    def report_failure(self, model: str):
        with self.lock:
            self.fail_counts[model] += 1
            if self.fail_counts[model] >= CIRCUIT_BREAKER_THRESHOLD:
                self.circuit_open_until[model] = time.time() + CIRCUIT_BREAKER_RESET


def call_llm(model: str, messages: list, rotator: ModelRotator = None) -> dict:
    """Call the LLM via OpenAI chat completions. Returns parsed response."""
    url = f"{PROXY_BASE}/v1/chat/completions"
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {PROXY_KEY}",
    }
    payload = {
        "model": model,
        "max_tokens": 4096,
        "messages": messages,
        "tools": SWARM_TOOLS,
        "tool_choice": "auto",
    }

    resp = requests.post(url, headers=headers, json=payload, timeout=API_TIMEOUT)

    if resp.status_code != 200:
        err_text = resp.text[:500]
        if rotator:
            rotator.report_failure(model)
        raise RuntimeError(f"HTTP {resp.status_code}: {err_text}")

    data = resp.json()
    if rotator:
        rotator.report_success(model)
    return data


# ─── Task Execution ─────────────────────────────────────────────────────────
GO_TOOL_TEMPLATE = """package tools

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "time"
)

// HandleXxx implements the xxx tool natively.
// Replaces: <github_url>
// Tools exposed: <list tool names>
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


def build_system_prompt(task: dict) -> str:
    """Build the system prompt for a worker."""
    return (
        "You are an autonomous coding agent working on the TormentNexus project.\n"
        f"WORKSPACE: {WORKSPACE}\n"
        f"GO_TOOLS_DIR: {GO_TOOLS_DIR}\n"
        f"REGISTRY_FILE: {REGISTRY_FILE}\n"
        "\n"
        "CRITICAL RULES:\n"
        "- You MUST use tools immediately. Never just describe what you would do — DO IT.\n"
        "- Your FIRST action must ALWAYS be a tool call (read_file, run_bash, grep_search).\n"
        "- Use read_file to examine existing code before making changes.\n"
        f"- Use run_bash to build/test: cd {WORKSPACE}/go && go build -buildvcs=false ./cmd/tormentnexus\n"
        "- Only modify registry.go if you need to register new tool handlers.\n"
        "- When done, call task_complete with a summary.\n"
        "\n"
        "GO TOOL PATTERN — every new Go tool file MUST follow this structure:\n"
        "```go\n" + GO_TOOL_TEMPLATE + "```\n"
        "\n"
        "Available helpers (already in parity.go — do NOT redefine):\n"
        "- getString(args, keys...) → (string, bool)\n"
        "- getInt(args, keys...) → int\n"
        "- getBool(args, key) → bool\n"
        "- ok(text string) → (ToolResponse, error)\n"
        "- err(text string) → (ToolResponse, error)\n"
        "\n"
        "Registration in registry.go — add inside registerAll():\n"
        '    r.handlers["tool_name"] = HandleXxx\n'
        "\n"
        "BUILD CHECK is MANDATORY after creating/modifying any Go file:\n"
        f"    cd {WORKSPACE}/go && go build -buildvcs=false ./cmd/tormentnexus\n"
        "\n"
        "If build fails, read the error, fix the code, and rebuild. Keep retrying until it compiles.\n"
    )


def build_user_prompt(task: dict) -> str:
    """Build the task-specific user prompt."""
    desc = task.get("description", "")
    task_id = task.get("id", "")
    github_url = task.get("github_url", "")
    score = task.get("score", 0)

    prompt = f"TASK: {task_id}\n\n{desc}"
    if github_url and "github.com" in github_url:
        prompt += f"\n\nSource repository: {github_url}"
        # Extract owner/repo for API access
        match = re.match(r"https://github\.com/([^/]+)/([^/\s]+)", github_url)
        if match:
            owner, repo = match.group(1), match.group(2)
            prompt += f"\n\nTo read the README, use: run_bash with command: curl -s 'https://api.github.com/repos/{owner}/{repo}/readme' | python3 -c \"import sys,json,base64; d=json.load(sys.stdin); print(base64.b64decode(d.get('content','')).decode('utf-8','ignore')[:5000])\""

    prompt += f"\n\nScore: {score}/100"
    prompt += f"\n\nFirst steps: 1) Read any existing Go files in {GO_TOOLS_DIR} to understand patterns. 2) Research the source if needed. 3) Implement. 4) Register. 5) Build check. 6) Call task_complete when done."
    return prompt


def execute_task(task: dict, rotator: ModelRotator, worker_id: str) -> dict:
    """Execute a single task with multi-turn LLM conversation."""
    task_id = task.get("id", "unknown")
    log_prefix = f"[{worker_id}|{task_id}]"
    model = rotator.next_model()
    print(f"{log_prefix} Starting (model={model})")

    system = build_system_prompt(task)
    user_msg = build_user_prompt(task)
    messages = [
        {"role": "system", "content": system},
        {"role": "user", "content": user_msg},
    ]

    total_api_calls = 0
    start_time = time.time()
    files_changed = []

    for turn in range(MAX_TURNS):
        # Pick model (rotate on retries)
        if turn > 0 and turn % 3 == 0:
            model = rotator.next_model()

        # Call LLM
        for attempt in range(MAX_RETRIES):
            try:
                data = call_llm(model, messages, rotator)
                total_api_calls += 1
                break
            except Exception as e:
                print(f"{log_prefix} API error (attempt {attempt + 1}): {str(e)[:120]}")
                if attempt < MAX_RETRIES - 1:
                    time.sleep(2**attempt)  # Exponential backoff
                    model = rotator.next_model()  # Try different model
                else:
                    return {
                        "status": "failed",
                        "error": str(e)[:500],
                        "api_calls": total_api_calls,
                        "turns": turn,
                        "elapsed": time.time() - start_time,
                        "model": model,
                    }

        # Parse response
        choice = data.get("choices", [{}])[0]
        msg = choice.get("message", {})
        finish = choice.get("finish_reason", "stop")
        tool_calls = msg.get("tool_calls", [])
        content = msg.get("content", "") or ""

        # Add assistant message to history
        assistant_msg = {"role": "assistant", "content": content}
        if tool_calls:
            assistant_msg["tool_calls"] = tool_calls
        messages.append(assistant_msg)

        # No tool calls — check if done
        if not tool_calls:
            if turn == 0:
                # Model just talked, didn't use tools — re-prompt
                messages.append(
                    {
                        "role": "user",
                        "content": "You must use tools to do the work. Start by calling read_file or run_bash now.",
                    }
                )
                continue

            # Natural end of conversation
            return {
                "status": "completed",
                "summary": content[:500],
                "api_calls": total_api_calls,
                "turns": turn + 1,
                "elapsed": time.time() - start_time,
                "model": model,
                "files_changed": files_changed,
            }

        # Process tool calls
        tool_results = []
        task_done = False
        task_result = {}

        for tc in tool_calls:
            fn = tc.get("function", {})
            tool_name = fn.get("name", "")
            try:
                tool_args = json.loads(fn.get("arguments", "{}"))
            except json.JSONDecodeError:
                tool_args = {}

            # Execute the tool
            result_text = execute_tool(tool_name, tool_args)

            # Check for task completion
            if result_text == "__TASK_COMPLETE__":
                task_done = True
                task_result = {
                    "status": "completed",
                    "summary": tool_args.get("summary", "Task completed"),
                    "files_changed": tool_args.get("files_changed", files_changed),
                    "discoveries": tool_args.get("discoveries", ""),
                    "api_calls": total_api_calls,
                    "turns": turn + 1,
                    "elapsed": time.time() - start_time,
                    "model": model,
                }
                result_text = "Task marked as complete."
            elif "write_file" in tool_name:
                files_changed.append(tool_args.get("path", ""))

            tool_results.append(
                {
                    "role": "tool",
                    "tool_call_id": tc.get("id", ""),
                    "content": result_text,
                }
            )

        # Add tool results to conversation
        messages.extend(tool_results)

        if task_done:
            return task_result

        # Log progress
        tool_names = [tc.get("function", {}).get("name", "?") for tc in tool_calls]
        elapsed = time.time() - start_time
        print(f"{log_prefix} Turn {turn + 1}: {tool_names} ({elapsed:.0f}s)")

    # Max turns exceeded
    return {
        "status": "max_turns",
        "summary": f"Reached {MAX_TURNS} turns without completing",
        "api_calls": total_api_calls,
        "turns": MAX_TURNS,
        "elapsed": time.time() - start_time,
        "model": model,
        "files_changed": files_changed,
    }


# ─── Task Generation ────────────────────────────────────────────────────────
def get_mcp_tasks(limit: int = 50) -> list:
    """Get pending MCP servers from state DB as tasks."""
    conn = sqlite3.connect(STATE_DB)
    rows = conn.execute(
        "SELECT name, github_url, score FROM mcp_servers "
        "WHERE status='pending' ORDER BY score DESC LIMIT ?",
        (limit,),
    ).fetchall()
    conn.close()

    tasks = []
    for name, url, score in rows:
        snake = re.sub(r"[^a-z0-9]", "_", name.lower())
        tasks.append(
            {
                "id": f"mcp-{name}",
                "description": (
                    f"Assimilate MCP server '{name}' as a native Go tool module.\n\n"
                    f"Steps:\n"
                    f"1. Read the source at {url} (use curl via run_bash to get README)\n"
                    f"2. Identify what tools/functions it exposes\n"
                    f"3. Create {GO_TOOLS_DIR}/{snake}.go following the GO TOOL PATTERN\n"
                    f"4. Register handlers in {REGISTRY_FILE}\n"
                    f"5. Build check: cd {WORKSPACE}/go && go build -buildvcs=false ./cmd/tormentnexus\n"
                    f"6. Fix any compilation errors until it builds clean"
                ),
                "github_url": url,
                "score": score,
            }
        )
    return tasks


def get_hermes_tasks(limit: int = 50) -> list:
    """Get pending Hermes addons from state DB as tasks."""
    conn = sqlite3.connect(STATE_DB)
    rows = conn.execute(
        "SELECT name, category, description FROM hermes_addons "
        "WHERE status='pending' LIMIT ?",
        (limit,),
    ).fetchall()
    conn.close()

    tasks = []
    for name, category, desc in rows:
        tasks.append(
            {
                "id": f"hermes-{name}",
                "description": (
                    f"Assimilate Hermes addon '{name}' (category: {category}).\n"
                    f"Description: {desc}\n\n"
                    f"Decision tree:\n"
                    f"- If it maps to an existing Go tool → skip (note which tool)\n"
                    f"- If it's prompt/instructions only → create a skill SKILL.md\n"
                    f"- If it needs API calls → implement as Go tool\n"
                    f"Build check required for any Go changes."
                ),
            }
        )
    return tasks


def get_go_improvement_tasks(limit: int = 50) -> list:
    """Generate tasks to improve existing Go tools (add descriptions, fix issues)."""
    tasks = []
    go_dir = Path(GO_TOOLS_DIR)

    for go_file in sorted(go_dir.glob("*.go")):
        if go_file.name.endswith("_test.go") or go_file.name == "registry.go":
            continue

        content = go_file.read_text(encoding="utf-8", errors="replace")

        # Check for missing descriptions in tool registrations
        if "description" not in content.lower() and "handle" in content.lower():
            tasks.append(
                {
                    "id": f"improve-desc-{go_file.stem}",
                    "description": (
                        f"Add proper descriptions to tool handlers in {go_file}.\n"
                        f"Read the file, understand what each Handle function does, "
                        f"and add a description string for each tool. "
                        f"Also check for any obvious bugs or missing error handling. "
                        f"Build check required."
                    ),
                }
            )

        if len(tasks) >= limit:
            break

    return tasks


# ─── Registry Lock ──────────────────────────────────────────────────────────
_registry_lock = threading.Lock()
_registry_lock_holder = None


def acquire_registry_lock(worker_id: str, timeout: float = 120) -> bool:
    """Acquire exclusive access to registry.go for modification."""
    global _registry_lock_holder
    acquired = _registry_lock.acquire(timeout=timeout)
    if acquired:
        _registry_lock_holder = worker_id
    return acquired


def release_registry_lock():
    """Release exclusive access to registry.go."""
    global _registry_lock_holder
    _registry_lock_holder = None
    _registry_lock.release()


# ─── State DB Updates ───────────────────────────────────────────────────────
def update_mcp_status(name: str, status: str, go_file: str = "", tools: str = "[]"):
    """Update MCP server status in state DB."""
    conn = sqlite3.connect(STATE_DB)
    conn.execute(
        "UPDATE mcp_servers SET status=?, go_file=?, tools_exposed=?, updated_at=CURRENT_TIMESTAMP WHERE name=?",
        (status, go_file, tools, name),
    )
    conn.commit()
    conn.close()


def update_hermes_status(
    name: str, status: str, go_file: str = "", skill_name: str = ""
):
    """Update Hermes addon status in state DB."""
    conn = sqlite3.connect(STATE_DB)
    conn.execute(
        "UPDATE hermes_addons SET status=?, go_file=?, skill_name=?, notes=CURRENT_TIMESTAMP WHERE name=?",
        (status, go_file, skill_name, name),
    )
    conn.commit()
    conn.close()


# ─── Worker ─────────────────────────────────────────────────────────────────
def worker(
    worker_id: str, task_queue: list, results: list, rotator: ModelRotator, stats: dict
):
    """Worker thread that processes tasks from the queue."""
    while True:
        # Get next task
        with threading.Lock():
            if not task_queue:
                break
            task = task_queue.pop(0)
            task_id = task.get("id", "unknown")
            stats["active"] += 1

        print(f"[{worker_id}] Picked up {task_id}")
        retry_count = task.get("_retry_count", 0)

        try:
            result = execute_task(task, rotator, worker_id)
            result["id"] = task_id
            result["worker"] = worker_id

            # If task needs registry modification, acquire lock
            if result.get("files_changed") and any(
                "registry.go" in f for f in result.get("files_changed", [])
            ):
                pass  # Registry changes happen inside the LLM tool calls

            # Update state DB based on task type
            if task_id.startswith("mcp-"):
                mcp_name = task_id[4:]
                status = "implemented" if result["status"] == "completed" else "failed"
                go_file = ""
                for f in result.get("files_changed", []):
                    if f.endswith(".go") and "registry" not in f:
                        go_file = f
                        break
                update_mcp_status(mcp_name, status, go_file)
            elif task_id.startswith("hermes-"):
                hermes_name = task_id[7:]
                status = "implemented" if result["status"] == "completed" else "pending"
                update_hermes_status(hermes_name, status)

            # Handle retries for failures
            if (
                result["status"] in ("failed", "max_turns")
                and retry_count < MAX_RETRIES
            ):
                task["_retry_count"] = retry_count + 1
                with threading.Lock():
                    task_queue.append(task)  # Re-queue
                print(
                    f"[{worker_id}] {task_id} failed, re-queued (retry {retry_count + 1})"
                )
            else:
                with threading.Lock():
                    results.append(result)
                    stats["done"] += 1
                    stats["active"] -= 1

            elapsed = result.get("elapsed", 0)
            turns = result.get("turns", 0)
            status = result.get("status", "?")
            print(f"[{worker_id}] {task_id}: {status} ({turns}t, {elapsed:.0f}s)")

        except Exception as e:
            print(f"[{worker_id}] {task_id} EXCEPTION: {str(e)[:200]}")
            traceback.print_exc()
            with threading.Lock():
                stats["active"] -= 1
                if retry_count < MAX_RETRIES:
                    task["_retry_count"] = retry_count + 1
                    task_queue.append(task)
                else:
                    results.append(
                        {
                            "id": task_id,
                            "status": "error",
                            "error": str(e)[:500],
                            "worker": worker_id,
                        }
                    )
                    stats["done"] += 1


# ─── Stats Dashboard ────────────────────────────────────────────────────────
def stats_reporter(stats: dict, task_queue: list, stop_event: threading.Event):
    """Print stats every 15 seconds."""
    while not stop_event.is_set():
        with threading.Lock():
            pending = len(task_queue)
            active = stats.get("active", 0)
            done = stats.get("done", 0)
            failed = stats.get("failed", 0)

        elapsed = time.time() - stats.get("start_time", time.time())
        rate = done / max(elapsed / 60, 0.1)
        print(
            f"[STATS] [{elapsed:.0f}s] pending:{pending} active:{active} done:{done} fail:{failed} | rate:{rate:.1f}/min"
        )
        stop_event.wait(15)


# ─── Main ────────────────────────────────────────────────────────────────────
def main():
    parser = argparse.ArgumentParser(description="TormentNexus Swarm")
    parser.add_argument(
        "--workers", type=int, default=3, help="Number of parallel workers"
    )
    parser.add_argument(
        "--track",
        choices=["mcp", "hermes", "improve"],
        default="mcp",
        help="Task track",
    )
    parser.add_argument("--limit", type=int, default=50, help="Max tasks to process")
    parser.add_argument("--tasks", type=str, help="JSON file with task definitions")
    parser.add_argument("--model", type=str, help="Force a specific model")
    args = parser.parse_args()

    # Load tasks
    if args.tasks:
        with open(args.tasks) as f:
            tasks = json.load(f)
        print(f"Loaded {len(tasks)} tasks from {args.tasks}")
    elif args.track == "mcp":
        tasks = get_mcp_tasks(args.limit)
        print(f"Loaded {len(tasks)} pending MCP servers from state DB")
    elif args.track == "hermes":
        tasks = get_hermes_tasks(args.limit)
        print(f"Loaded {len(tasks)} pending Hermes addons from state DB")
    elif args.track == "improve":
        tasks = get_go_improvement_tasks(args.limit)
        print(f"Loaded {len(tasks)} Go improvement tasks")

    if not tasks:
        print("No tasks to process!")
        return

    # Override models if specified
    global MODELS
    if args.model:
        MODELS = [args.model]

    print(f"\n{'=' * 60}")
    print("  TORMENTNEXUS SWARM")
    print(f"  Workers: {args.workers}")
    print(f"  Models: {MODELS}")
    print(f"  Tasks: {len(tasks)}")
    print(f"  Track: {args.track}")
    print(f"  Proxy: {PROXY_BASE}")
    print(f"{'=' * 60}\n")

    # Setup
    rotator = ModelRotator(MODELS)
    results = []
    stats = {"active": 0, "done": 0, "failed": 0, "start_time": time.time()}
    stop_event = threading.Event()

    # Start stats reporter
    stats_thread = threading.Thread(
        target=stats_reporter, args=(stats, tasks, stop_event), daemon=True
    )
    stats_thread.start()

    # Start workers
    with ThreadPoolExecutor(max_workers=args.workers) as executor:
        futures = []
        for i in range(args.workers):
            worker_id = f"W{i + 1}"
            future = executor.submit(worker, worker_id, tasks, results, rotator, stats)
            futures.append(future)

        # Wait for all workers
        for future in as_completed(futures):
            try:
                future.result()
            except Exception as e:
                print(f"Worker crashed: {e}")

    stop_event.set()

    # Summary
    completed = [r for r in results if r.get("status") == "completed"]
    failed = [r for r in results if r.get("status") != "completed"]
    total_time = time.time() - stats["start_time"]

    print(f"\n{'=' * 60}")
    print("  SWARM COMPLETE")
    print(f"  Completed: {len(completed)}/{len(results)}")
    print(f"  Failed: {len(failed)}")
    print(f"  Total time: {total_time:.0f}s ({total_time / 60:.1f}min)")
    print(f"  Rate: {len(completed) / max(total_time / 60, 0.1):.1f} tasks/min")
    print(f"{'=' * 60}")

    # Save results
    results_file = Path(WORKSPACE) / "swarm_results.json"
    results_file.write_text(
        json.dumps(
            {
                "summary": {
                    "total": len(results),
                    "completed": len(completed),
                    "failed": len(failed),
                    "total_time_s": total_time,
                },
                "completed_tasks": completed,
                "failed_tasks": failed,
            },
            indent=2,
        )
    )
    print(f"Results saved to {results_file}")

    # Print failed tasks for debugging
    if failed:
        print("\nFailed tasks:")
        for r in failed[:20]:
            print(f"  {r.get('id', '?')}: {r.get('error', r.get('summary', ''))[:120]}")


if __name__ == "__main__":
    main()
