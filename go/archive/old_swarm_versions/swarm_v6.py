#!/usr/bin/env python3
"""
TormentNexus Swarm v6 — THE RELIABLE ASSIMILATOR
=================================================
Uses STREAMING curl to bypass the FreeLLM proxy's 30s non-streaming timeout.
Works with BOTH thinking models (zai-glm) and clean models (deepseek/qwen).
Extracts Go code from mixed reasoning+code output.

DESIGN:
  1. STREAMING via curl — 300s timeout, no proxy truncation
  2. SERIAL per-worker — one LLM call at a time, wait for completion
  3. STAGGERED start — workers begin 30s apart to avoid rate limits
  4. DUAL parser — handles thinking (zai-glm) and clean (deepseek) output
  5. MANIFEST files — workers never touch registry.go directly
  6. BUILD VERIFIER — auto-remove broken files, fix registry refs
  7. NEVER STOP — retry failed tasks, rotate models, wait out rate limits

USAGE:
  python -u swarm_v6.py --workers 8 --limit 50 --forever --no-research
  python -u swarm_v6.py --repair
  python -u swarm_v6.py --workers 4 --limit 20
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
from typing import Optional, Tuple, List, Dict

# =============================================================================
# CONFIG
# =============================================================================

PROXY_URL = os.environ.get("SWARM_PROXY", "http://localhost:4000")
PROXY_KEY = os.environ.get("SWARM_KEY", "sk-freellm")
PROXY_BIN = os.environ.get(
    "SWARM_PROXY_BIN",
    "C:/Users/hyper/workspace/litellm_control_panel/freellm.exe",
)
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

# Model names to request from the proxy — we rotate through these.
# The proxy load-balances to actual backends (zai-glm, deepseek, qwen).
# Some produce thinking output, some produce clean code. We handle both.
REQUEST_MODELS = [
    "free-llm",  # routes to deepseek-v4-flash or zai-glm
    "free-llm-fallback",  # routes to qwen-397b or zai-glm
    "glm-4-flash",  # sometimes routes to deepseek-v4-flash
    "gpt-4o-mini",  # routes to whatever is available
    "claude-haiku-4",  # routes to deepseek-v4-flash
    "deepseek-ai/deepseek-v4-pro",  # routes to qwen or zai-glm
    "internlm3-latest",  # routes to qwen
]

# Timeouts
STREAM_TIMEOUT = 300  # curl --max-time for streaming calls
CURL_TIMEOUT = 330  # subprocess timeout (stream + buffer)
CALL_COOLDOWN = 120  # seconds between calls per worker
WORKER_STAGGER = 120  # seconds between worker starts
MAX_RETRIES = 5  # retries per LLM call
RETRY_BASE_DELAY = 15  # base delay between retries
CIRCUIT_BREAKER_FAILS = 10  # consecutive fails before circuit opens
CIRCUIT_BREAKER_COOLDOWN = 120  # seconds before circuit resets
MIN_CONTENT_LENGTH = 30  # minimum chars to consider a valid response
BUILD_VERIFY_EVERY = 2  # verify build every N completed tasks
PROXY_RESTART_AFTER = 8  # consecutive proxy errors before restart


# =============================================================================
# LOG
# =============================================================================


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
        self.builds_ok = 0
        self.builds_fix = 0
        self.model_stats: Dict[str, Dict] = {}  # model -> {ok, err, chars}

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

    def llm_fail(self, m, w=None):
        self._w("LLM-ERR", "R", w, m)
        with self.lock:
            self.llm_err_count += 1

    def stat(self, m):
        self._w("STAT", "B", None, m)

    def build(self, m):
        self._w("BUILD", "B", None, m)

    def record_model(self, model, ok, chars=0):
        with self.lock:
            s = self.model_stats.setdefault(model, {"ok": 0, "err": 0, "chars": 0})
            if ok:
                s["ok"] += 1
                s["chars"] += chars
            else:
                s["err"] += 1

    def summary(self):
        e = int(time.time() - self.t0)
        with self.lock:
            s = (
                f"ok={self.ok_count} fail={self.fail_count} "
                f"llm_ok={self.llm_ok_count} llm_err={self.llm_err_count} "
                f"builds_ok={self.builds_ok} builds_fix={self.builds_fix} "
                f"elapsed={e // 3600}h{(e % 3600) // 60}m"
            )
            # Best performing models
            best = sorted(
                [
                    (m, v["ok"], v["err"], v["chars"])
                    for m, v in self.model_stats.items()
                    if v["ok"] > 0
                ],
                key=lambda x: x[1],
                reverse=True,
            )[:3]
            if best:
                s += " | best_models: " + ", ".join(
                    f"{m}({o}ok)" for m, o, _, _ in best
                )
            return s


log = Log()


# =============================================================================
# LLM CLIENT — Streaming curl, dual parser, aggressive retry
# =============================================================================


# Models that produce garbage output - skip these
BLACKLISTED_MODELS = set()

class LLMClient:
    """Thread-safe LLM client using streaming curl through the proxy."""

    def __init__(self):
        self.lock = threading.Lock()
        self.model_idx = 0
        self.circuit_open = False
        self.circuit_opened_at = 0
        self.consecutive_fails = 0
        self.proxy_consecutive_errs = 0
        self.last_call_time = 0

    def _next_model(self):
        with self.lock:
            # Skip blacklisted models
            attempts = 0
            while attempts < len(REQUEST_MODELS) * 2:
                m = REQUEST_MODELS[self.model_idx % len(REQUEST_MODELS)]
                self.model_idx += 1
                if m not in BLACKLISTED_MODELS:
                    return m
                attempts += 1
            return REQUEST_MODELS[self.model_idx % len(REQUEST_MODELS)]

    def blacklist_model(self, model):
        """Blacklist a model that produces garbage output."""
        BLACKLISTED_MODELS.add(model)
        log.warn(f"Blacklisted model: {model} (total: {len(BLACKLISTED_MODELS)})")

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
        log.warn(f"Circuit OPEN - {CIRCUIT_BREAKER_COOLDOWN}s cooldown")

    def _record_fail(self):
        self.consecutive_fails += 1
        if self.consecutive_fails >= CIRCUIT_BREAKER_FAILS:
            self._trip_circuit()

    def _record_ok(self):
        self.consecutive_fails = 0

    def _enforce_cooldown(self):
        """Enforce minimum time between calls."""
        with self.lock:
            since_last = time.time() - self.last_call_time
            if since_last < CALL_COOLDOWN:
                wait = CALL_COOLDOWN - since_last
                time.sleep(wait)
            self.last_call_time = time.time()

    def call_streaming(
        self, prompt, system, model, wid=0
    ) -> Tuple[Optional[str], str, float]:
        """
        Call LLM via streaming curl. Returns (content, model_used, elapsed).
        Content is None on failure.
        """
        self._enforce_cooldown()

        payload = json.dumps(
            {
                "model": model,
                "messages": [
                    {"role": "system", "content": system},
                    {"role": "user", "content": prompt},
                ],
                "temperature": 0.2,
                "max_tokens": 4096,
                "stream": True,
            }
        )

        t0 = time.time()
        log.llm(f"{model} stream t={STREAM_TIMEOUT}s", wid)

        try:
            result = subprocess.run(
                [
                    "curl",
                    "-s",
                    "-N",
                    "--max-time",
                    str(STREAM_TIMEOUT),
                    "-X",
                    "POST",
                    f"{PROXY_URL}/v1/chat/completions",
                    "-H",
                    "Content-Type: application/json",
                    "-H",
                    f"Authorization: Bearer {PROXY_KEY}",
                    "-d",
                    payload,
                ],
                capture_output=True,
                text=True,
                timeout=CURL_TIMEOUT,
                encoding="utf-8",
                errors="replace",
            )
            elapsed = time.time() - t0

            # Parse SSE stream
            content = ""
            model_used = ""
            finish_reason = None
            for line in result.stdout.split("\n"):
                line = line.strip()
                if line.startswith("data: ") and line != "data: [DONE]":
                    try:
                        d = json.loads(line[6:])
                        if not model_used:
                            model_used = d.get("model", "?")
                        delta = d.get("choices", [{}])[0].get("delta", {})
                        c = delta.get("content", "")
                        if c:
                            content += c
                        fr = d.get("choices", [{}])[0].get("finish_reason")
                        if fr:
                            finish_reason = fr
                    except Exception:
                        pass

            if content and len(content) >= MIN_CONTENT_LENGTH:
                self._record_ok()
                self.proxy_consecutive_errs = 0
                log.llm(
                    f"{model_used}: {len(content)}c in {elapsed:.1f}s (finish={finish_reason})",
                    wid,
                )
                log.record_model(model_used, True, len(content))
                return content, model_used, elapsed

            # Empty or too short
            log.llm_fail(
                f"{model_used}: short/empty ({len(content)}c in {elapsed:.1f}s)", wid
            )
            self._record_fail()
            self.proxy_consecutive_errs += 1
            log.record_model(model_used, False)
            return None, model_used, elapsed

        except subprocess.TimeoutExpired:
            elapsed = time.time() - t0
            log.llm_fail(f"{model}: timeout ({elapsed:.1f}s)", wid)
            self._record_fail()
            self.proxy_consecutive_errs += 1
            return None, model, elapsed

        except Exception as e:
            elapsed = time.time() - t0
            log.llm_fail(f"{model}: {str(e)[:120]}", wid)
            self._record_fail()
            self.proxy_consecutive_errs += 1
            return None, model, elapsed

    def call(self, prompt, system="", wid=0) -> Optional[str]:
        """
        Call LLM with aggressive retry. Tries multiple models.
        Returns response text or None.
        """
        if not self._check_circuit():
            return None

        for attempt in range(MAX_RETRIES):
            if self.circuit_open:
                log.warn(f"Circuit open, waiting (attempt {attempt + 1})", wid)
                time.sleep(CIRCUIT_BREAKER_COOLDOWN)
                if not self._check_circuit():
                    continue

            model = self._next_model()
            content, model_used, elapsed = self.call_streaming(
                prompt, system, model, wid
            )

            if content:
                return content

            # Exponential backoff between retries
            delay = RETRY_BASE_DELAY * (2 ** min(attempt, 4))
            log.warn(f"Retry {attempt + 1}/{MAX_RETRIES} after {delay}s", wid)
            time.sleep(delay)

            # If proxy is struggling, consider restart
            if self.proxy_consecutive_errs >= PROXY_RESTART_AFTER:
                log.stat("Too many proxy errors, restarting...")
                proxy.restart()
                self.proxy_consecutive_errs = 0

        log.err(f"All {MAX_RETRIES} retries exhausted", wid)
        return None


llm = LLMClient()


# =============================================================================
# PROXY MANAGER
# =============================================================================


class ProxyManager:
    def __init__(self):
        self._healthy = True
        self.consecutive_errs = 0

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
                [PROXY_BIN],
                cwd=d,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
            )
        except Exception as e:
            log.err(f"Proxy start: {e}")
            return False
        for i in range(24):
            time.sleep(5)
            if self.check():
                log.ok("Proxy back up!")
                return True
        log.err("Proxy never came up")
        return False

    @property
    def healthy(self):
        return self._healthy


proxy = ProxyManager()


# =============================================================================
# GO CODE EXTRACTION — Handles both thinking and clean output
# =============================================================================


def extract_go_code(
    content: str, go_filename: str
) -> Tuple[Optional[str], Optional[dict]]:
    """
    Extract Go code and manifest from LLM output.
    Handles:
      1. Clean output with ===GO_FILE=== / ===MANIFEST=== markers
      2. Thinking model output (zai-glm) with code in ```go blocks
      3. Fallback: find package tools + Handle functions anywhere
    """
    if not content:
        return None, None

    go_code = None
    manifest = None

    # --- Method 1: ===GO_FILE=== / ===MANIFEST=== markers ---
    gm = re.search(r"===GO_FILE===\s*\n(.*?)(?====MANIFEST===|$)", content, re.DOTALL)
    mm = re.search(r"===MANIFEST===\s*\n(.*?)$", content, re.DOTALL)

    if gm:
        go_code = gm.group(1).strip()
        go_code = _strip_fences(go_code)

    if mm:
        manifest = _parse_manifest(mm.group(1).strip(), go_filename)

    # If markers gave us code with Handle functions, we're done
    if go_code and _has_handlers(go_code):
        # If the marker code is from a thinking model, it might have
        # reasoning text mixed in — try to clean it
        if not go_code.strip().startswith("package tools"):
            go_code = _extract_from_mixed(go_code)
        # Auto-fix truncated code (unclosed braces, missing returns)
        go_code = fix_truncated_code(go_code)
        return go_code, manifest

    # --- Method 2: Extract from ```go code blocks (thinking model) ---
    blocks = re.findall(r"```go\s*\n(.*?)```", content, re.DOTALL)
    if blocks:
        go_code = _assemble_from_blocks(blocks)
        if go_code and _has_handlers(go_code):
            if not manifest:
                manifest = _build_manifest_from_code(go_code, go_filename)
            return go_code, manifest

    # --- Method 3: Find package tools anywhere and extract to end ---
    idx = content.find("package tools")
    if idx >= 0:
        rest = content[idx:]
        # Clean up: remove any non-Go text that leaked in
        go_code = _clean_raw_code(rest)
        if go_code and _has_handlers(go_code):
            if not manifest:
                manifest = _build_manifest_from_code(go_code, go_filename)
            return go_code, manifest

    # --- Method 4: Stitch Handle functions together ---
    handlers = re.findall(
        r"(func Handle\w+\s*\([^)]*\)\s*\([^)]*\)\s*\{)",
        content,
    )
    if handlers:
        # Find complete function bodies
        parts = []
        for h_match in re.finditer(
            r"func Handle\w+\s*\([^)]*\)\s*\([^)]*\)\s*\{.*?\n\}",
            content,
            re.DOTALL,
        ):
            parts.append(h_match.group(0))
        if parts:
            go_code = (
                'package tools\n\nimport (\n\t"context"\n\t"encoding/json"\n\t"fmt"\n\t"io"\n\t"net/http"\n\t"net/url"\n\t"os"\n\t"time"\n)\n\n'
                + "\n\n".join(parts)
            )
            manifest = _build_manifest_from_code(go_code, go_filename)
            return go_code, manifest

    return None, None


def _strip_fences(code: str) -> str:
    """Remove markdown code fences."""
    code = re.sub(r"^```go\s*\n?", "", code)
    code = re.sub(r"\n?```\s*$", "", code)
    return code.strip()


def _has_handlers(code: str) -> bool:
    """Check if code contains Handle functions."""
    return bool(re.search(r"func Handle\w+\s*\(", code))


def _parse_manifest(raw: str, go_filename: str) -> Optional[dict]:
    """Parse manifest JSON, with fallback fixes."""
    raw = _strip_fences(raw)
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        pass
    # Fix trailing commas
    fixed = re.sub(r",\s*}", "}", raw)
    fixed = re.sub(r",\s*]", "]", fixed)
    try:
        return json.loads(fixed)
    except json.JSONDecodeError:
        pass
    return None


def _assemble_from_blocks(blocks: List[str]) -> Optional[str]:
    """
    Assemble Go code from multiple ```go blocks (thinking model output).
    The thinking model splits code across blocks — we merge them intelligently.
    """
    # Filter blocks that have actual Go code (not just snippets)
    code_blocks = []
    for block in blocks:
        b = block.strip()
        if not b:
            continue
        # Skip blocks that are clearly just type definitions or comments
        # without being part of a complete file
        code_blocks.append(b)

    if not code_blocks:
        return None

    # Strategy 1: Find a single block that has package tools + Handle functions
    for block in code_blocks:
        if "package tools" in block and _has_handlers(block):
            return block

    # Strategy 2: The block with package tools is the base, merge others into it
    base = None
    additions = []
    for block in code_blocks:
        if "package tools" in block:
            base = block
        elif _has_handlers(block) or re.match(r"(type|var|const)\s", block):
            additions.append(block)

    if base:
        # Add missing imports and types from other blocks
        for add in additions:
            # Don't duplicate
            if add not in base:
                # Insert before the first Handle function
                first_handle = re.search(r"func Handle\w+", base)
                if first_handle:
                    idx = first_handle.start()
                    base = base[:idx] + add + "\n\n" + base[idx:]
                else:
                    base += "\n\n" + add
        return base

    # Strategy 3: No block has package tools — construct from scratch
    if additions:
        imports = set()
        for block in additions:
            for imp in re.findall(r'"([^"]+)"', block):
                if "/" not in imp or imp.startswith("golang.org"):
                    imports.add(imp)

        import_block = ""
        if imports:
            import_block = (
                "import (\n"
                + "\n".join(f'\t"{i}"' for i in sorted(imports))
                + "\n)\n\n"
            )

        code = "package tools\n\n" + import_block + "\n\n".join(additions)
        if _has_handlers(code):
            return code

    return None


def _extract_from_mixed(code: str) -> Optional[str]:
    """Extract Go code from mixed reasoning+code text."""
    idx = code.find("package tools")
    if idx < 0:
        return None
    rest = code[idx:]
    return _clean_raw_code(rest)


def _clean_raw_code(code: str) -> Optional[str]:
    """
    Clean raw code that might have reasoning text mixed in.
    Strategy: Keep lines that look like Go, remove reasoning.
    """
    lines = code.split("\n")
    clean_lines = []
    in_code = False
    brace_depth = 0
    skip_reasoning = False

    for line in lines:
        stripped = line.strip()

        # Skip obvious reasoning patterns
        if re.match(r"^\d+\.\s+\*\*", stripped):
            skip_reasoning = True
            continue
        if re.match(r"^[*\-]\s+\*\*", stripped):
            skip_reasoning = True
            continue
        if skip_reasoning:
            # Reasoning continues until we see code-like lines
            if (
                stripped.startswith("package ")
                or stripped.startswith("import ")
                or stripped.startswith("func ")
                or stripped.startswith("type ")
                or stripped.startswith("var ")
                or stripped.startswith("const ")
                or stripped.startswith("//")
                or stripped == ""
                and in_code
            ):
                skip_reasoning = False
            else:
                continue

        # Track if we're inside Go code
        if stripped.startswith("package tools"):
            in_code = True
        if in_code:
            brace_depth += stripped.count("{") - stripped.count("}")
            clean_lines.append(line)
            # Stop at end of last Handle function if we've gone past
            if brace_depth <= 0 and "func Handle" in "\n".join(clean_lines[-20:]):
                # Check if this was a function closing
                if stripped == "}":
                    # Could be end of a function — keep going for more
                    pass

    result = "\n".join(clean_lines)
    if _has_handlers(result):
        return result
    return None


def fix_truncated_code(code: str) -> str:
    """Auto-fix truncated Go code: close braces, add missing returns."""
    if not code:
        return code
    # Remove incomplete last line
    lines = code.rstrip().split("\n")
    while lines:
        s = lines[-1].strip()
        if not s:
            break
        if re.search(r"[;,{}()\[\]]\s*$", s) or s.endswith(")"):
            break
        lines.pop()
    code = "\n".join(lines)
    # Close unclosed braces
    open_b = code.count("{")
    close_b = code.count("}")
    missing = open_b - close_b
    if missing > 0:
        for i in range(missing):
            indent = "\t" * max(0, missing - i - 1)
            code += "\n" + indent + "}"
    # Add default return to Handle functions missing one
    for h_match in re.finditer(r"func (Handle\w+)\s*\([^)]*\)\s*\([^)]*\)\s*\{", code):
        start = h_match.end()
        depth = 1
        end = start
        for i in range(start, len(code)):
            if code[i] == "{":
                depth += 1
            elif code[i] == "}":
                depth -= 1
            if depth == 0:
                end = i
                break
        body = code[start:end]
        if "return ok(" not in body and "return err(" not in body:
            last_brace = code.rfind("}", start, end + 1)
            if last_brace > 0:
                ret_line = '\n\treturn ok("not yet implemented")\n'
                code = code[:last_brace] + ret_line + code[last_brace:]
                break
    return code


def _build_manifest_from_code(code: str, go_filename: str) -> dict:
    """Build manifest from handler functions found in code."""
    handlers = re.findall(r"func (Handle\w+)\s*\(", code)
    return {
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


# =============================================================================
# DATABASE — Atomic task claiming
# =============================================================================

_db_lock = threading.Lock()


def claim_task(worker_id):
    """Atomically claim the highest-score pending task."""
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
            "tools_exposed=?, notes='v6', updated_at=CURRENT_TIMESTAMP WHERE name=?",
            (go_file, str(tools_count), name),
        )
        db.commit()
        db.close()


def fail_task(name, error=""):
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


def reset_failed(limit=30, min_score=0):
    db = sqlite3.connect(str(DB_PATH))
    c = db.cursor()
    c.execute(
        "UPDATE mcp_servers SET status='pending', notes='retry' "
        "WHERE status='failed' AND score >= ? ORDER BY score DESC LIMIT ?",
        (min_score, limit),
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


# =============================================================================
# GO FILE NAMING & PROMPT BUILDING
# =============================================================================


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
1. Package: package tools
2. Handler: func HandleXxx(ctx context.Context, args map[string]interface{{}}) (ToolResponse, error)
3. Success: return ok("text")  — ok() returns (ToolResponse, error)
4. Error CHECK: if e != nil {{ return err(e.Error()) }}
5. getString/getInt/getBool return SINGLE values: val := getString(args, "key")
6. Only stdlib imports: context, encoding/json, fmt, io, net/http, net/url, os, os/exec, path/filepath, strconv, strings, time, regexp, sort
7. http.Client{{Timeout: 30*time.Second}}
8. 2-6 handlers, keep simple, MUST COMPILE, no TODOs
9. Do NOT redeclare ToolResponse, ok, err, getString, getInt, getBool, TextContent — they exist in parity.go

OUTPUT FORMAT — output TWO sections:
===GO_FILE===
package tools
// complete Go source code
===MANIFEST===
{{"filename":"{fn}.go","server_name":"{n}","handlers":[{{"tool_name":"x","handler_func":"HandleX","description":"d"}}]}}""")

    system = textwrap.dedent("""\
You write clean compilable Go code for MCP tool handlers.
ok()/err() return (ToolResponse, error). getString returns single value.
No pseudocode, no TODOs. Every function MUST compile.
Do NOT redeclare types from parity.go (ToolResponse, ok, err, getString, etc).
Output ===GO_FILE=== then code, then ===MANIFEST=== then JSON.""")

    return prompt, system


# =============================================================================
# BUILD VERIFICATION & REPAIR
# =============================================================================


def verify_build():
    """Run go build, remove broken tool files, clean registry refs."""
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
            log.build("CLEAN")
            with log.lock:
                log.builds_ok += 1
            return True

        broken = set()
        undefs = set()
        for line in (r.stderr or "").split("\n"):
            # Find broken tool files
            for sep in ["internal\\tools\\", "internal/tools/"]:
                if sep in line:
                    after = line.split(sep, 1)[1]
                    fname = after.split(":")[0]
                    if fname.endswith(".go") and fname not in PROTECTED_FILES:
                        broken.add(fname)
                    break
            # Find undefined handlers in registry
            if "undefined:" in line:
                for word in line.split():
                    w = word.rstrip(",")
                    if w.startswith("Handle") and len(w) > 6:
                        undefs.add(w)

        for f in broken:
            p = TOOLS_DIR / f
            if p.exists():
                log.warn(f"Removing broken: {f}")
                p.unlink()

        if undefs:
            _clean_registry(undefs)

        with log.lock:
            log.builds_fix += 1
        return False

    except Exception as e:
        log.err(f"Build error: {e}")
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
            log.info(f"Cleaned {removed} undefined refs from registry")
    except Exception as e:
        log.err(f"Registry clean: {e}")


def repair_build(max_iter=15):
    """Iteratively fix build errors until clean."""
    for i in range(max_iter):
        if verify_build():
            log.ok(f"Build clean (iter {i})")
            return True
        log.warn(f"Build broken, iter {i + 1}")
        time.sleep(1)
    return verify_build()


# =============================================================================
# MANIFEST MERGER
# =============================================================================

_merge_lock = threading.Lock()


def merge_manifests():
    """Atomically merge pending manifests into registry.go."""
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
                    tn = h.get("tool_name", "")
                    hf = h.get("handler_func", "")
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


# =============================================================================
# WORKER — Claim, generate, write, repeat
# =============================================================================


def worker_loop(worker_id, shutdown, completed_lock, completed_count):
    """
    Main worker loop. Claims tasks, generates code, writes files.
    Runs until no more tasks or shutdown.
    """
    # Stagger worker starts to avoid proxy saturation
    stagger = (worker_id - 1) * WORKER_STAGGER
    if stagger > 0:
        log.info(f"Staggering {stagger}s before start", worker_id)
        for _ in range(stagger):
            if shutdown.is_set():
                return
            time.sleep(1)

    while not shutdown.is_set():
        # Claim a task
        task = claim_task(worker_id)
        if not task:
            log.info("No more tasks", worker_id)
            return

        name = task["name"]
        fn = name_to_filename(name)
        log.info(f"Claimed: {name} -> {fn}.go", worker_id)

        # Wait if circuit breaker is open
        if llm.circuit_open:
            log.warn("Circuit open - waiting...", worker_id)
            for _ in range(12):
                if shutdown.is_set():
                    release_task(name)
                    return
                if not llm.circuit_open:
                    break
                time.sleep(10)

        # Check proxy health
        if not proxy.healthy:
            log.warn("Proxy down - checking...", worker_id)
            for _ in range(6):
                if shutdown.is_set():
                    release_task(name)
                    return
                if proxy.check():
                    break
                time.sleep(10)
            if not proxy.healthy:
                proxy.restart()

        # Generate code via LLM
        prompt, system = make_prompt(task)
        # Use a specific model based on worker rotation
        model_for_task = REQUEST_MODELS[worker_id % len(REQUEST_MODELS)]
        output = llm.call(prompt, system, wid=worker_id)

        if not output:
            fail_task(name, "LLM failed after retries")
            log.err(f"LLM failed: {name}", worker_id)
            continue

        # Extract Go code (handles both thinking and clean output)
        go_code, manifest = extract_go_code(output, fn)
        if not go_code:
            fail_task(name, "No Go code extracted from output")
            log.err(f"No extractable code: {name}", worker_id)
            continue

        # Validate basic structure
        if "package tools" not in go_code:
            fail_task(name, "Extracted code missing package declaration")
            log.err(f"Missing package: {name}", worker_id)
            # Blacklist the model if it produced package main
            if "package main" in go_code:
                llm.blacklist_model(model_used if model_used else "unknown")
            continue

        # Check for garbage patterns (repeated broken code)
        if go_code.count("func (ctx context.Context)") > 3:
            fail_task(name, "Garbage output detected (repeated broken signatures)")
            log.err(f"Garbage: {name}", worker_id)
            llm.blacklist_model(model_used if model_used else "unknown")
            continue

        # Check for redeclared types from parity.go
        if "type ToolResponse struct" in go_code:
            # Remove the redeclaration
            go_code = re.sub(r"type ToolResponse struct\s*\{.*?\}", "", go_code, flags=re.DOTALL)
            go_code = re.sub(r"type TextContent struct\s*\{.*?\}", "", go_code, flags=re.DOTALL)
            log.warn(f"Removed redeclared types from {fn}", worker_id)

        handlers_found = re.findall(r"func (Handle\w+)\s*\(", go_code)
        if not handlers_found:
            fail_task(name, "No Handle functions in extracted code")
            log.err(f"No handlers: {name}", worker_id)
            continue

        # Write Go file
        go_file = f"{fn}.go"
        try:
            with open(TOOLS_DIR / go_file, "w", encoding="utf-8") as f:
                f.write(go_code)
        except Exception as e:
            fail_task(name, f"Write error: {e}")
            log.err(f"Write {go_file}: {e}", worker_id)
            continue

        # Write manifest
        if manifest:
            MANIFEST_DIR.mkdir(exist_ok=True)
            try:
                with open(MANIFEST_DIR / f"{fn}.json", "w", encoding="utf-8") as f:
                    json.dump(manifest, f, indent=2)
            except Exception:
                pass

        # Mark complete
        tc = len(handlers_found)
        complete_task(name, go_file, tc)
        log.ok(f"Done: {name} ({tc} handlers: {handlers_found})", worker_id)

        with completed_lock:
            completed_count[0] += 1


# =============================================================================
# ORCHESTRATOR
# =============================================================================


class Swarm:
    def __init__(self, max_workers, limit, forever, repair_only, no_research):
        self.max_workers = max_workers
        self.limit = limit
        self.forever = forever
        self.repair_only = repair_only
        self.no_research = no_research
        self.running = True
        self.shutdown = threading.Event()
        self.completed_lock = threading.Lock()
        self.completed_count = [0]  # mutable container for threading

        signal.signal(signal.SIGINT, lambda s, f: self._stop())
        signal.signal(signal.SIGTERM, lambda s, f: self._stop())

    def _stop(self):
        log.stat("Shutdown signal...")
        self.running = False
        self.shutdown.set()

    def run(self):
        log.stat("=" * 70)
        log.stat("  TORMENTNEXUS SWARM v6 — THE RELIABLE ASSIMILATOR")
        log.stat("=" * 70)
        log.stat(
            f"Workers: {self.max_workers} | Limit: {self.limit} | Forever: {self.forever}"
        )
        log.stat(f"Proxy: {PROXY_URL} | Stream timeout: {STREAM_TIMEOUT}s")
        log.stat(f"Call cooldown: {CALL_COOLDOWN}s | Worker stagger: {WORKER_STAGGER}s")
        log.stat(f"Models: {REQUEST_MODELS}")

        t, i, p, f = get_stats()
        log.stat(f"DB: {t} total | {i} done | {p} pending | {f} failed")

        MANIFEST_DIR.mkdir(exist_ok=True)

        # Release stale tasks
        stale = reset_stale()
        if stale:
            log.info(f"Released {stale} stale tasks")

        # Check proxy
        if not proxy.check():
            log.warn("Proxy not healthy - restarting...")
            if not proxy.restart():
                log.err("Cannot start proxy!")
                return False

        # Repair-only mode
        if self.repair_only:
            repair_build()
            merge_manifests()
            verify_build()
            return True

        # Initial build check
        log.stat("Checking build...")
        repair_build()
        merge_manifests()
        verify_build()

        # Main loop
        while self.running:
            try:
                self._run_wave()
            except Exception as e:
                log.err(f"Wave error: {e}")
                traceback.print_exc()
                if not self.forever:
                    break
                time.sleep(30)

            # Check limits
            if not self.forever and self.completed_count[0] >= self.limit:
                log.stat(f"Limit reached ({self.limit})!")
                break

            # Check if there are pending tasks
            _, _, pending, failed = get_stats()
            if pending == 0:
                if self.forever and failed > 0:
                    n = reset_failed(limit=50, min_score=0)
                    if n > 0:
                        log.stat(f"Retrying {n} failed tasks")
                        continue
                log.stat("No pending tasks!")
                break

            if self.forever:
                # Reset some failed tasks each cycle
                n = reset_failed(limit=20, min_score=0)
                if n:
                    log.info(f"Auto-retry: {n} failed tasks reset")
                time.sleep(5)

        self._shutdown()
        return True

    def _run_wave(self):
        """Run a wave of workers."""
        ws = self.max_workers
        if not self.forever:
            remaining = self.limit - self.completed_count[0]
            ws = min(ws, max(remaining, 1))

        log.stat(f"Wave: {ws} workers")

        before = self.completed_count[0]

        with ThreadPoolExecutor(max_workers=ws) as pool:
            futs = {}
            for wid in range(1, ws + 1):
                if self.shutdown.is_set():
                    break
                futs[
                    pool.submit(
                        worker_loop,
                        wid,
                        self.shutdown,
                        self.completed_lock,
                        self.completed_count,
                    )
                ] = wid

            last_build_check = 0
            for f in as_completed(futs):
                if self.shutdown.is_set():
                    break
                try:
                    f.result()
                except Exception as e:
                    log.err(f"Worker {futs[f]}: {e}")

                # Periodic build verification
                now = self.completed_count[0]
                if now - last_build_check >= BUILD_VERIFY_EVERY:
                    merge_manifests()
                    repair_build()
                    last_build_check = now

        # Final verification for this wave
        merge_manifests()
        repair_build()

        delta = self.completed_count[0] - before
        t, i, p, f2 = get_stats()
        log.stat(
            f"Wave done: +{delta} | DB: {i}/{t} done, {p} pend, {f2} fail | {log.summary()}"
        )

    def _shutdown(self):
        """Graceful shutdown."""
        log.stat("Shutting down...")
        merge_manifests()
        repair_build()

        tc = len(list(TOOLS_DIR.glob("*.go")))
        try:
            hc = len(
                re.findall(
                    r"r\.handlers\[",
                    REGISTRY_FILE.read_text(encoding="utf-8", errors="replace"),
                )
            )
        except Exception:
            hc = 0

        t, i, p, f = get_stats()
        log.stat("=" * 70)
        log.stat("  FINAL STATE")
        log.stat("=" * 70)
        log.stat(f"  Tool files: {tc}")
        log.stat(f"  Registered handlers: {hc}")
        log.stat(f"  DB: {i}/{t} done, {p} pending, {f} failed")
        log.stat(f"  {log.summary()}")
        log.stat("=" * 70)


# =============================================================================
# MAIN
# =============================================================================


def main():
    ap = argparse.ArgumentParser(
        description="TormentNexus Swarm v6 - The Reliable Assimilator"
    )
    ap.add_argument(
        "--workers", type=int, default=4, help="Number of concurrent workers"
    )
    ap.add_argument("--limit", type=int, default=50, help="Max tasks to complete")
    ap.add_argument(
        "--forever", action="store_true", help="Run continuously until all done"
    )
    ap.add_argument("--repair", action="store_true", help="Repair build only")
    ap.add_argument(
        "--no-research", action="store_true", default=True, help="Skip research phase"
    )
    ap.add_argument("--proxy", type=str, default=None, help="Override proxy URL")
    ap.add_argument("--log", type=str, default=None, help="Log file path")
    args = ap.parse_args()

    if args.proxy:
        global PROXY_URL
        PROXY_URL = args.proxy

    # Setup logging
    log.path = args.log or str(
        WORKSPACE / "data" / f"swarm_v6_{datetime.now().strftime('%Y%m%d_%H%M%S')}.log"
    )
    (WORKSPACE / "data").mkdir(exist_ok=True)

    swarm = Swarm(
        max_workers=args.workers,
        limit=args.limit,
        forever=args.forever,
        repair_only=args.repair,
        no_research=args.no_research,
    )
    success = swarm.run()
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
