"""Runtime trace ingestion infrastructure.

Phase 0 shipped the schema, the redaction chokepoint, and the symbol
resolver. Phase 1 added the OTel JSON file-import path. Phase 2 surfaced
``_runtime_confidence`` on five existing tools. Phase 3 added three new
runtime-aware MCP tools. Phase 4 added SQL query log ingest +
``runtime_columns`` table for dbt-style data layers. Phase 5 (this) adds
application-log / stack-frame ingest + ``runtime_stack_events`` table
for severity-tagged error-frequency tracking.

Public surface:
- redact_trace_record(record, source) — single redaction chokepoint
- resolve_to_symbol_id(conn, file_path, line_no, function_name) — best-effort resolver
- ingest_otel_file(...) — orchestrate parse → redact → resolve → upsert (OTel)
- ingest_sql_log_file(...) — orchestrate parse → redact → resolve → upsert (SQL log)
- parse_otel_file(path) — pure OTel JSON / JSONL iterator (no DB writes)
- parse_sql_log_file(path) — pure pg_stat_statements / JSON-Lines SQL log iterator
- VALID_SOURCES — frozenset of accepted source labels
"""

from .confidence import (
    RuntimeConfidenceProbe,
    attach_runtime_confidence,
    attach_runtime_confidence_by_file,
)
from .ingest import ingest_otel_file, ingest_otel_stream
from .otel import OtelSpan, iter_otel_from_text, parse_otel_file
from .redact import redact_trace_record
from .resolve import resolve_to_symbol_id
from .sql_ingest import ingest_sql_log_file, ingest_sql_log_stream
from .sql_log import SqlQueryRecord, iter_sql_from_text, parse_sql_log_file
from .stack_ingest import ingest_stack_log_file, ingest_stack_log_stream
from .stack_log import StackEvent, StackFrame, iter_stack_from_text, parse_stack_log_file

VALID_SOURCES = frozenset({"otel", "sql_log", "stack_log", "apm"})

__all__ = [
    "redact_trace_record",
    "resolve_to_symbol_id",
    "parse_otel_file",
    "iter_otel_from_text",
    "ingest_otel_file",
    "ingest_otel_stream",
    "OtelSpan",
    "parse_sql_log_file",
    "iter_sql_from_text",
    "ingest_sql_log_file",
    "ingest_sql_log_stream",
    "SqlQueryRecord",
    "parse_stack_log_file",
    "iter_stack_from_text",
    "ingest_stack_log_file",
    "ingest_stack_log_stream",
    "StackEvent",
    "StackFrame",
    "RuntimeConfidenceProbe",
    "attach_runtime_confidence",
    "attach_runtime_confidence_by_file",
    "VALID_SOURCES",
]
