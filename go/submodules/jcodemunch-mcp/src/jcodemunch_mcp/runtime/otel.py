"""OTLP JSON file parser for runtime trace ingest (Phase 1).

Reads the file format produced by the OTel Collector ``file`` exporter
(or any tool that emits OTLP/JSON), iterates spans, and extracts the
code-location attributes defined by the OpenTelemetry semantic
conventions:

  - ``code.filepath``  — file path (absolute or repo-relative)
  - ``code.lineno``    — 1-based line number within ``code.filepath``
  - ``code.function``  — function / method name
  - ``code.namespace`` — module / class namespace (kept for diagnostics)

Spans without any code attribute are still iterated and reported as
unmapped — the orchestrator records them in ``runtime_unmapped`` for
diagnostic purposes.

Supported file shapes:
  - **JSON-Lines** (one resource-span object per line) — the OTel
    Collector ``file`` exporter default.
  - **JSON array** at the top level — common when a tool dumps via
    ``json.dump(spans, f)``.
  - **Single top-level object** with ``resourceSpans``.

Gzipped (``.gz``) variants are detected by extension and decompressed
transparently.
"""

from __future__ import annotations

import gzip
import json
import logging
from dataclasses import dataclass, field
from pathlib import Path
from typing import Iterable, Iterator, Optional

logger = logging.getLogger(__name__)


@dataclass
class OtelSpan:
    """Extracted code-location data for one OTel span.

    Only the fields that drive symbol resolution + the duration metric
    are surfaced; everything else (trace id, span id, full attribute
    bag) is dropped at parse time so the redaction chokepoint never has
    to scrub it.
    """
    name: str
    file_path: Optional[str]
    line_no: Optional[int]
    function_name: Optional[str]
    namespace: Optional[str]
    duration_ms: Optional[float]
    # The raw attribute bag is preserved for the redaction layer to scan;
    # callers must hand it to ``redact_trace_record()`` before any
    # storage call. Keys/values are stringified at parse time so the
    # redaction module's regexes are sufficient.
    extra: dict = field(default_factory=dict)


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def parse_otel_file(path: str) -> Iterator[OtelSpan]:
    """Yield one OtelSpan per span in the file.

    Args:
        path: Filesystem path. Reads gzip-compressed if extension is .gz.
    """
    p = Path(path)
    if not p.exists():
        raise FileNotFoundError(f"OTel trace file not found: {path}")
    opener = gzip.open if p.suffix == ".gz" else open
    with opener(p, "rt", encoding="utf-8") as fh:  # type: ignore[operator]
        text = fh.read()
    yield from iter_otel_from_text(text)


def iter_otel_from_text(text: str) -> Iterator[OtelSpan]:
    """Yield OtelSpan instances from an in-memory OTLP text payload.

    The shared ``parse_otel_file`` and Phase 6 HTTP routes both delegate
    here so the wire format and the file format are decoded by exactly
    the same logic.

    Args:
        text: OTLP/JSON content. May be JSON-Lines, a single
            top-level object, or a JSON array of resourceSpans records.

    Raises:
        ValueError: when the payload begins with a byte that's neither
            ``[`` nor ``{``. Empty / whitespace-only input is yielded
            silently as zero spans (caller treats that as a no-op).
    """
    # Find the first non-whitespace byte to decide the layout.
    idx = 0
    n = len(text)
    while idx < n and text[idx].isspace():
        idx += 1
    if idx >= n:
        return
    first = text[idx]

    if first == "[":
        try:
            payload = json.loads(text)
        except json.JSONDecodeError as e:
            raise ValueError(f"OTel JSON array parse failed: {e}") from e
        for record in payload:
            yield from _extract_spans_from_record(record)
        return

    if first == "{":
        # Could be either JSON-Lines (each line is one record) or a
        # single big object. Try the single-object case first; fall
        # back to JSON-Lines on multi-document detection.
        try:
            payload = json.loads(text)
            yield from _extract_spans_from_record(payload)
            return
        except json.JSONDecodeError:
            for line in text.splitlines():
                line = line.strip()
                if not line:
                    continue
                try:
                    record = json.loads(line)
                except json.JSONDecodeError:
                    logger.debug("Skipping malformed OTel JSONL line", exc_info=True)
                    continue
                yield from _extract_spans_from_record(record)
            return

    raise ValueError(f"Unsupported OTel file shape: starts with {first!r}")


# ---------------------------------------------------------------------------
# Private — span extraction
# ---------------------------------------------------------------------------


def _extract_spans_from_record(record: dict) -> Iterable[OtelSpan]:
    """Walk a single OTel/JSON record (typically one ``resourceSpans``
    bag from the file exporter) and yield ``OtelSpan`` per span.
    """
    if not isinstance(record, dict):
        return
    # OTLP layout: resourceSpans -> scopeSpans -> spans
    resource_spans = record.get("resourceSpans") or [record]
    if not isinstance(resource_spans, list):
        return
    for rs in resource_spans:
        if not isinstance(rs, dict):
            continue
        scope_spans = rs.get("scopeSpans") or rs.get("instrumentationLibrarySpans") or []
        if not isinstance(scope_spans, list):
            continue
        for ss in scope_spans:
            if not isinstance(ss, dict):
                continue
            spans = ss.get("spans") or []
            if not isinstance(spans, list):
                continue
            for span in spans:
                parsed = _span_to_otelspan(span)
                if parsed is not None:
                    yield parsed


def _span_to_otelspan(span: dict) -> Optional[OtelSpan]:
    """Project one OTLP span to ``OtelSpan``, dropping non-code attributes."""
    if not isinstance(span, dict):
        return None
    attrs = _flatten_attributes(span.get("attributes") or [])
    file_path = _coerce_str(attrs.get("code.filepath"))
    line_no = _coerce_int(attrs.get("code.lineno"))
    function_name = _coerce_str(attrs.get("code.function"))
    namespace = _coerce_str(attrs.get("code.namespace"))
    name = _coerce_str(span.get("name")) or ""

    duration_ms: Optional[float] = None
    start = span.get("startTimeUnixNano")
    end = span.get("endTimeUnixNano")
    try:
        if start is not None and end is not None:
            duration_ms = (int(end) - int(start)) / 1_000_000.0
    except (TypeError, ValueError):
        duration_ms = None

    # Only surface non-code attributes in `extra` so the redactor scans
    # the smallest possible payload. The code.* attributes are already
    # captured structurally above.
    extra = {k: v for k, v in attrs.items() if not k.startswith("code.")}
    return OtelSpan(
        name=name,
        file_path=file_path,
        line_no=line_no,
        function_name=function_name,
        namespace=namespace,
        duration_ms=duration_ms,
        extra=extra,
    )


# ---------------------------------------------------------------------------
# OTLP attribute decoding
# ---------------------------------------------------------------------------


def _flatten_attributes(attrs: list) -> dict:
    """Decode the OTLP attribute list shape ``[{key, value: {<typeKey>: ...}}]``.

    OTLP wraps every value in a one-key dict whose key encodes the type
    (``stringValue``, ``intValue``, ``boolValue``, ``doubleValue``,
    ``arrayValue``, ``kvlistValue``). For ingest we coerce everything
    to a string scalar — the redactor operates on strings, and the
    structural fields we care about (file/line/function) are never
    arrays.
    """
    out: dict = {}
    if not isinstance(attrs, list):
        return out
    for item in attrs:
        if not isinstance(item, dict):
            continue
        key = item.get("key")
        if not isinstance(key, str):
            continue
        out[key] = _otlp_value(item.get("value"))
    return out


def _otlp_value(v) -> Optional[str]:
    """Reduce an OTLP value object to a single string (or None)."""
    if v is None:
        return None
    if isinstance(v, (str, int, float, bool)):
        return str(v)
    if not isinstance(v, dict):
        return None
    for key in ("stringValue", "intValue", "doubleValue", "boolValue"):
        if key in v:
            return str(v[key])
    if "arrayValue" in v:
        arr = v["arrayValue"].get("values") if isinstance(v["arrayValue"], dict) else None
        if isinstance(arr, list):
            return ",".join(str(_otlp_value(x) or "") for x in arr)
    return None


def _coerce_str(value) -> Optional[str]:
    if value is None:
        return None
    if isinstance(value, str):
        return value or None
    return str(value)


def _coerce_int(value) -> Optional[int]:
    if value is None:
        return None
    try:
        return int(value)
    except (TypeError, ValueError):
        return None
