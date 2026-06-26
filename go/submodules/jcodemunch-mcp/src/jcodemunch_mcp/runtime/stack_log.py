"""Phase 5 application-log / stack-frame parser.

Pure parsing layer (no DB writes). Yields ``StackEvent`` instances —
one per logical exception/log entry, with the parsed frame chain and
inferred severity.

Three stack-trace dialects are recognised:

1. **Python tracebacks.** ``Traceback (most recent call last):``
   followed by alternating ``File "...", line N, in <name>`` /
   indented source lines, terminated by an exception line such as
   ``ValueError: ...``. Frames are extracted in source order
   (innermost-first → outermost-last; we record the chain as written).

2. **JVM tracebacks.** ``com.example.Foo: message`` (or just
   ``Exception: ...``) followed by ``    at com.example.Foo.bar(File.java:42)``
   lines. The parser captures the method-qualified name as the function
   and ``File.java`` as the filename. ``Caused by:`` chains are flattened
   into one event with all frames.

3. **Node.js stacks.** ``Error: message`` followed by
   ``    at funcName (path/to/file.js:42:7)`` (or the anonymous
   ``    at path/to/file.js:42:7`` form). The trailing column number
   is dropped; we only track ``(file, line, function)``.

Severity is heuristic from the log line that introduces the trace:
``ERROR | FATAL | CRITICAL`` → ``error``;
``WARN | WARNING`` → ``warn``;
anything else (or no preceding tag) → ``info``. The default is
``info`` because some pipelines log every exception at info-level for
post-hoc triage.

Two file shapes are supported:

* **Plain text application log.** One stack per occurrence, scanned
  line-by-line. Multiple stacks per file is the common case.

* **JSON-Lines structured log.** ``{"message": "...",
  "stack_trace": "...", "severity": "ERROR", "ts": "..."}``. The
  ``stack_trace`` field contains the formatted trace; ``severity``
  / ``level`` overrides the heuristic when present.

Reusable for Phase 6 (HTTP live-ingest endpoint will hand records
straight to the parser).
"""

from __future__ import annotations

import gzip
import io
import json
import logging
import re
from dataclasses import dataclass, field
from pathlib import Path
from typing import Iterator, Optional

logger = logging.getLogger(__name__)


# ---------------------------------------------------------------------------
# Public types
# ---------------------------------------------------------------------------


VALID_SEVERITIES = ("error", "warn", "info")


@dataclass
class StackFrame:
    """A single parsed (file, line, function) triple."""
    file_path: str
    line_no: Optional[int]
    function_name: Optional[str]


@dataclass
class StackEvent:
    """One logical stack — every frame contributes a runtime_stack_events row."""
    severity: str
    """One of ``VALID_SEVERITIES``."""

    frames: list[StackFrame] = field(default_factory=list)
    """Frames in source order (innermost-first for Python, outermost-first
    for JVM/Node.js mirrors the language convention)."""

    timestamp: Optional[str] = None
    """ISO-8601 timestamp from the log line, when available."""

    message: Optional[str] = None
    """Truncated exception text; never persisted as-is — redaction
    chokepoint runs over this before any storage write."""


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def parse_stack_log_file(path: str) -> Iterator[StackEvent]:
    """Yield one StackEvent per stack found in ``path``.

    Format dispatch:
      * ``.jsonl`` / ``.json`` — JSON-Lines structured log (each record
        with a ``stack_trace`` field). Top-level array fallback honoured.
      * any other extension — plain-text scan. ``.gz`` transparent.

    Empty / stack-free files yield nothing.
    """
    p = Path(path)
    if not p.exists():
        raise FileNotFoundError(f"Stack log file not found: {path}")

    inner = p.suffix.lower()
    is_gz = inner == ".gz"
    if is_gz:
        inner = Path(p.stem).suffix.lower()

    opener = (lambda: gzip.open(p, "rt", encoding="utf-8", errors="replace")) if is_gz else (
        lambda: open(p, "r", encoding="utf-8", errors="replace")
    )

    if inner in (".jsonl", ".json", ".ndjson"):
        yield from _parse_json_lines(opener)
    else:
        yield from _parse_plain_text(opener)


def iter_stack_from_text(text: str, *, fmt: str = "auto") -> Iterator[StackEvent]:
    """Yield StackEvent from an in-memory stack-log payload.

    ``fmt`` selects the parser:
      * ``'auto'`` (default) — heuristic: if the first non-whitespace
        char is ``{`` or ``[`` treat as JSON-Lines / array; otherwise
        plain-text.
      * ``'plain'`` — force the line-by-line traceback scanner.
      * ``'jsonl'`` — force the JSON-Lines path.
    """
    if not text.strip():
        return

    if fmt == "auto":
        head = text.lstrip()
        fmt = "jsonl" if head[:1] in ("{", "[") else "plain"

    if fmt == "jsonl":
        yield from _parse_json_lines(lambda: io.StringIO(text))
    else:
        yield from iter_events_from_text(text)


# ---------------------------------------------------------------------------
# JSON-Lines structured log
# ---------------------------------------------------------------------------


def _parse_json_lines(opener) -> Iterator[StackEvent]:
    with opener() as f:
        first = f.read(4096)
    stripped = first.lstrip()
    if stripped.startswith("["):
        # Top-level array — slurp, decode, walk
        with opener() as f2:
            data = f2.read()
        try:
            payload = json.loads(data)
        except json.JSONDecodeError as exc:
            logger.warning("Stack log: top-level array decode failed: %s", exc)
            return
        if not isinstance(payload, list):
            return
        for obj in payload:
            ev = _event_from_json_obj(obj)
            if ev is not None:
                yield ev
        return

    with opener() as f3:
        for raw in f3:
            line = raw.strip()
            if not line or line.startswith("#"):
                continue
            try:
                obj = json.loads(line)
            except json.JSONDecodeError:
                continue
            ev = _event_from_json_obj(obj)
            if ev is not None:
                yield ev


def _event_from_json_obj(obj) -> Optional[StackEvent]:
    if not isinstance(obj, dict):
        return None
    trace = obj.get("stack_trace") or obj.get("stack") or obj.get("exception")
    if not isinstance(trace, str) or not trace.strip():
        return None

    severity_field = obj.get("severity") or obj.get("level") or obj.get("loglevel")
    severity = _normalise_severity(severity_field) if severity_field else _detect_severity(trace[:256])

    ts = obj.get("ts") or obj.get("timestamp") or obj.get("time")
    msg = obj.get("message")

    frames = _parse_frames(trace)
    if not frames:
        return None
    return StackEvent(
        severity=severity,
        frames=frames,
        timestamp=ts if isinstance(ts, str) else None,
        message=msg if isinstance(msg, str) else None,
    )


# ---------------------------------------------------------------------------
# Plain-text application log
# ---------------------------------------------------------------------------

_PY_TRACEBACK_HEADER = re.compile(r"^Traceback \(most recent call last\):\s*$")
_JVM_AT_LINE = re.compile(r"^\s*at\s+([\w$.<>]+)\(([^):]+):(\d+)\)\s*$")
_JVM_HEADER = re.compile(r"^([\w.$]+(?:Exception|Error|Throwable))(?::\s*(.*))?\s*$")
_NODE_HEADER = re.compile(r"^([\w]*Error)(?::\s*(.*))?\s*$")
# Node.js stack frames take two shapes:
#   ``    at handleRequest (src/server.js:120:7)``    — named, parens
#   ``    at src/anon.js:9:5``                       — anonymous, no parens
# The location is allowed to contain ``:`` so that node-builtin module IDs like
# ``node:events`` survive the parse. Strategy: capture the location lazily up
# to the trailing ``:line(:col)?`` suffix.
_NODE_AT_LINE = re.compile(
    r"^\s*at\s+"
    r"(?:(?P<func>[^(]+?)\s+\()?"
    r"(?P<file>.+?)"
    r":(?P<line>\d+)"
    r"(?::\d+)?"
    r"\)?\s*$"
)
_LOG_LEVEL_HINT = re.compile(r"\b(FATAL|CRITICAL|ERROR|WARN(?:ING)?|INFO|DEBUG|TRACE)\b", re.IGNORECASE)
_TIMESTAMP_HINT = re.compile(
    r"\b(\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:[.,]\d+)?(?:Z|[+\-]\d{2}:?\d{2})?)\b"
)


def _parse_plain_text(opener) -> Iterator[StackEvent]:
    """Scan a plain-text log file for Python / JVM / Node.js stacks."""
    with opener() as f:
        text = f.read()
    yield from iter_events_from_text(text)


def iter_events_from_text(text: str) -> Iterator[StackEvent]:
    """Public helper so tests can feed in a string directly."""
    lines = text.splitlines()
    i = 0
    n = len(lines)
    while i < n:
        line = lines[i]
        # Python traceback?
        if _PY_TRACEBACK_HEADER.match(line):
            ev, j = _parse_python_block(lines, i)
            if ev is not None:
                yield ev
            i = j
            continue
        # JVM-shaped exception line followed by `at ...`?
        m_jvm = _JVM_HEADER.match(line.strip())
        if m_jvm and i + 1 < n and _JVM_AT_LINE.match(lines[i + 1]):
            ev, j = _parse_jvm_block(lines, i)
            if ev is not None:
                yield ev
            i = j
            continue
        # Node.js-shaped error line followed by `at ...`?
        m_node = _NODE_HEADER.match(line.strip())
        if m_node and i + 1 < n and _NODE_AT_LINE.match(lines[i + 1]):
            ev, j = _parse_node_block(lines, i)
            if ev is not None:
                yield ev
            i = j
            continue
        i += 1


def _parse_python_block(lines: list[str], start: int) -> tuple[Optional[StackEvent], int]:
    """Walk from a ``Traceback (most recent ...)`` header forward.

    Frames are alternating ``  File "...", line N, in <name>`` / source-line
    pairs. The block ends at the first non-``  File`` non-source-indented
    line that doesn't fit the pattern — typically the exception summary.
    """
    severity = _detect_severity_around(lines, start)
    timestamp = _detect_timestamp_around(lines, start)
    frames: list[StackFrame] = []
    i = start + 1
    n = len(lines)
    while i < n:
        m = re.match(r'^\s*File "([^"]+)", line (\d+), in (.+?)\s*$', lines[i])
        if m:
            frames.append(StackFrame(
                file_path=m.group(1),
                line_no=int(m.group(2)),
                function_name=m.group(3).strip(),
            ))
            # The next non-empty line is the source preview — skip it.
            i += 2
            continue
        # Exception summary line (e.g. ``ValueError: ...``) — end of block.
        stripped = lines[i].strip()
        if stripped and not stripped.startswith("File ") and i > start + 1:
            return (
                StackEvent(
                    severity=severity,
                    frames=frames,
                    timestamp=timestamp,
                    message=stripped,
                ) if frames else None,
                i + 1,
            )
        i += 1
    if not frames:
        return None, i
    return StackEvent(severity=severity, frames=frames, timestamp=timestamp), i


def _parse_jvm_block(lines: list[str], start: int) -> tuple[Optional[StackEvent], int]:
    """Walk from a ``com.example.Exception: ...`` header forward.

    Each frame is ``    at pkg.Class.method(File.java:N)``. The block ends
    when an indented line stops matching. ``Caused by:`` chains are
    swallowed (frames keep accumulating into one StackEvent).
    """
    severity = _detect_severity_around(lines, start)
    timestamp = _detect_timestamp_around(lines, start)
    header = lines[start].strip()
    frames: list[StackFrame] = []
    i = start + 1
    n = len(lines)
    while i < n:
        m = _JVM_AT_LINE.match(lines[i])
        if m:
            method_qual = m.group(1)
            file_name = m.group(2)
            line_no = int(m.group(3))
            function_name = method_qual.rsplit(".", 1)[-1]
            frames.append(StackFrame(
                file_path=file_name, line_no=line_no, function_name=function_name,
            ))
            i += 1
            continue
        # Tolerate ``Caused by:`` / ``... 5 more`` continuations
        stripped = lines[i].strip()
        if stripped.startswith("Caused by:") or stripped.startswith("..."):
            i += 1
            continue
        break
    if not frames:
        return None, i
    return (
        StackEvent(severity=severity, frames=frames, timestamp=timestamp, message=header),
        i,
    )


def _parse_node_block(lines: list[str], start: int) -> tuple[Optional[StackEvent], int]:
    """Walk a Node.js error header + ``at`` lines forward."""
    severity = _detect_severity_around(lines, start)
    timestamp = _detect_timestamp_around(lines, start)
    header = lines[start].strip()
    frames: list[StackFrame] = []
    i = start + 1
    n = len(lines)
    while i < n:
        m = _NODE_AT_LINE.match(lines[i])
        if m:
            func = (m.group("func") or "").strip() or None
            file_name = m.group("file").strip()
            try:
                line_no = int(m.group("line"))
            except ValueError:
                line_no = None
            frames.append(StackFrame(
                file_path=file_name, line_no=line_no, function_name=func,
            ))
            i += 1
            continue
        break
    if not frames:
        return None, i
    return (
        StackEvent(severity=severity, frames=frames, timestamp=timestamp, message=header),
        i,
    )


# ---------------------------------------------------------------------------
# Severity / timestamp helpers
# ---------------------------------------------------------------------------


def _normalise_severity(value: object) -> str:
    if not isinstance(value, str):
        return "info"
    v = value.strip().upper()
    if v in ("ERROR", "FATAL", "CRITICAL", "SEVERE"):
        return "error"
    if v in ("WARN", "WARNING"):
        return "warn"
    return "info"


def _detect_severity(text: str) -> str:
    m = _LOG_LEVEL_HINT.search(text)
    if not m:
        return "info"
    return _normalise_severity(m.group(1))


def _detect_severity_around(lines: list[str], idx: int, lookback: int = 3) -> str:
    """Look at the current line + a few lines above for a log-level tag.
    Plain-text app logs typically prefix the stack with the level on the
    line that introduces the exception."""
    start = max(0, idx - lookback)
    for j in range(idx, start - 1, -1):
        m = _LOG_LEVEL_HINT.search(lines[j])
        if m:
            return _normalise_severity(m.group(1))
    return "info"


def _detect_timestamp_around(lines: list[str], idx: int, lookback: int = 3) -> Optional[str]:
    start = max(0, idx - lookback)
    for j in range(idx, start - 1, -1):
        m = _TIMESTAMP_HINT.search(lines[j])
        if m:
            return m.group(1)
    return None


def _parse_frames(trace_text: str) -> list[StackFrame]:
    """Best-effort frame extraction from any of the three dialects.

    Used by the JSON-Lines path where a single record carries a fully
    formatted trace string. Tries Python first, then JVM, then Node.
    """
    frames: list[StackFrame] = []
    # Python style
    for m in re.finditer(r'File "([^"]+)", line (\d+), in (.+?)(?:\s|$)', trace_text):
        frames.append(StackFrame(
            file_path=m.group(1),
            line_no=int(m.group(2)),
            function_name=m.group(3).strip(),
        ))
    if frames:
        return frames

    # JVM style
    for m in _JVM_AT_LINE.finditer(trace_text):
        method_qual = m.group(1)
        file_name = m.group(2)
        line_no = int(m.group(3))
        frames.append(StackFrame(
            file_path=file_name,
            line_no=line_no,
            function_name=method_qual.rsplit(".", 1)[-1],
        ))
    if frames:
        return frames

    # Node style
    for m in _NODE_AT_LINE.finditer(trace_text):
        func = (m.group("func") or "").strip() or None
        try:
            line_no = int(m.group("line"))
        except (TypeError, ValueError):
            line_no = None
        frames.append(StackFrame(
            file_path=m.group("file").strip(),
            line_no=line_no,
            function_name=func,
        ))
    return frames
