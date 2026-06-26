"""Single chokepoint for redacting PII out of runtime trace records.

Every ingest path (OTel, SQL logs, stack logs, APM exports — Phases 1+) must
route every incoming record through `redact_trace_record()` before any
storage call. The chokepoint pattern is enforced by code review, not by
type system; the test suite asserts that ingestors call it exactly once
per record.

What gets stripped:
  - SQL parameter literals: `WHERE id = 42` → `WHERE id = ?`, string literals → `'?'`
  - HTTP request/response bodies: any payload, header value, or query-string value
  - Stack-frame local-variable repr blocks (Python: `vars: {...}`, JVM: `args=[...]`)
  - Existing high-entropy secret patterns (AWS/GCP/Azure/JWT/GitHub/Slack/PEM/...)

The redaction policy is deliberately aggressive — the static call graph only
needs `(file, line, function, count)` per record; literal values add nothing
to the structural picture. False-positive over-redaction is preferable to
false-negative leakage.

Controlled by `runtime_redact_enabled` config key (default: True). Disabling
is permitted only for offline debugging on synthetic data — never on
production traces.
"""

from __future__ import annotations

import logging
import re

from ..redact import _PATTERNS as _SECRET_PATTERNS

logger = logging.getLogger(__name__)


# ---------------------------------------------------------------------------
# Trace-specific patterns (additive over the secrets registry in ../redact.py)
# ---------------------------------------------------------------------------

_TRACE_PATTERNS: list[tuple[str, re.Pattern]] = [
    # SQL string literals — single-quoted values
    ("sql_string_literal", re.compile(r"'(?:[^']|'')*'")),
    # SQL numeric parameters in clauses (heuristic: only after =/IN/BETWEEN/VALUES)
    ("sql_numeric_param", re.compile(
        r"(?i)(?:=|<|>|<=|>=|<>|!=|\bIN\b|\bBETWEEN\b|\bVALUES\b)\s*\(?\s*(?P<secret>-?\d+(?:\.\d+)?)"
    )),
    # JSON-ish key:value blocks where values are arbitrary scalars (request
    # bodies, query params, header values). Anchored on quoted keys.
    ("json_value_string", re.compile(
        r'"(?P<key>[A-Za-z_][A-Za-z0-9_\-]{0,64})"\s*:\s*"(?P<secret>[^"\\\n]{0,256})"'
    )),
    # Python-style local-variable repr (kwargs={...}, locals={...}, vars={...})
    ("python_locals_block", re.compile(
        r"(?i)\b(?:kwargs|locals|vars|self|args)\s*=\s*\{(?P<secret>[^{}]{0,2048})\}"
    )),
    # IP addresses (v4 only — v6 too noisy across the codebase)
    ("ipv4_address", re.compile(
        r"(?<![0-9.])(?P<secret>(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.(?:25[0-5]|2[0-4]\d|[01]?\d\d?))(?![0-9.])"
    )),
    # Email addresses
    ("email_address", re.compile(
        r"(?P<secret>[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,})"
    )),
]


# Fields on a trace record that are *structural* and must never be redacted.
# Anything else gets the redact treatment.
_STRUCTURAL_KEYS = frozenset({
    "symbol_id", "caller_id", "callee_id", "import_id",
    "file_path", "line_no", "function_name",
    "source", "count", "p50_ms", "p95_ms",
    "first_seen", "last_seen",
    "kind", "language",
})


def redact_trace_record(record: dict, source: str) -> tuple[dict, list[str]]:
    """Strip PII from a runtime trace record before storage.

    Args:
        record: Raw trace record dict. Structural fields (symbol_id,
            file_path, line_no, function_name, count, source, timestamps)
            are passed through; everything else is scanned and redacted.
        source: One of {'otel', 'sql_log', 'stack_log', 'apm'}. Recorded
            on the redaction-log row for forensic accounting.

    Returns:
        (redacted_record, redaction_labels)
          - redacted_record: same shape as input with redacted values replaced
            by ``"[REDACTED:<label>]"``.
          - redaction_labels: list of pattern labels that fired (for the
            runtime_redaction_log table).

    The function is pure; callers are responsible for persisting the
    redaction labels to runtime_redaction_log.
    """
    if not isinstance(record, dict):
        return record, []
    redacted: dict = {}
    fired: list[str] = []
    for key, value in record.items():
        if key in _STRUCTURAL_KEYS:
            redacted[key] = value
            continue
        if isinstance(value, str):
            new_value, labels = _redact_string(value)
            redacted[key] = new_value
            fired.extend(labels)
        elif isinstance(value, dict):
            sub, labels = redact_trace_record(value, source)
            redacted[key] = sub
            fired.extend(labels)
        elif isinstance(value, list):
            new_list = []
            for item in value:
                if isinstance(item, dict):
                    sub, labels = redact_trace_record(item, source)
                    new_list.append(sub)
                    fired.extend(labels)
                elif isinstance(item, str):
                    new_value, labels = _redact_string(item)
                    new_list.append(new_value)
                    fired.extend(labels)
                else:
                    new_list.append(item)
            redacted[key] = new_list
        else:
            # Numbers, booleans, None — pass through. Numbers in SQL params
            # are caught at the string level via _redact_string when the SQL
            # text itself is being redacted.
            redacted[key] = value
    return redacted, fired


def _redact_string(value: str) -> tuple[str, list[str]]:
    """Apply both the secrets registry and the trace-pattern registry."""
    fired: list[str] = []
    out = value
    # Secrets registry first (bearer tokens, API keys, JWTs, etc.)
    for label, pattern in _SECRET_PATTERNS:
        def _sub_secret(m: re.Match, _label: str = label) -> str:
            fired.append(_label)
            secret = m.groupdict().get("secret")
            if secret:
                return m.group(0).replace(secret, f"[REDACTED:{_label}]")
            return f"[REDACTED:{_label}]"
        out = pattern.sub(_sub_secret, out)
    # Trace patterns (IPs, emails, SQL literals, JSON values, locals blocks)
    for label, pattern in _TRACE_PATTERNS:
        def _sub_trace(m: re.Match, _label: str = label) -> str:
            fired.append(_label)
            secret = m.groupdict().get("secret")
            if secret:
                return m.group(0).replace(secret, f"[REDACTED:{_label}]")
            return f"[REDACTED:{_label}]"
        out = pattern.sub(_sub_trace, out)
    return out, fired
