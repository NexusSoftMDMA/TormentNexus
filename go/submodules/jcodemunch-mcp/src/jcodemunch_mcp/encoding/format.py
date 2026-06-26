"""MUNCH on-wire format primitives.

MUNCH is jcodemunch's purpose-built compact encoding for tool responses.
A payload has four sections separated by blank lines:

    #MUNCH/1 tool=<name> enc=<encoding_id>
    @1=<path_prefix>
    $1=<symbol_fqn>

    key1=value1 key2=value2

    c,<col1>,<col2>,<col3>
    c,<col1>,<col2>,<col3>

Sections (all optional except header):
    1. Header  — version + tool + encoding id
    2. Legends — @N for path prefixes, $N for symbol fqns
    3. Scalars — whitespace-separated key=value pairs
    4. Tables  — CSV-style rows, first char is a table tag

Booleans encode as T/F. Null is an empty field. Values containing comma,
equals, whitespace, or a leading quote are double-quoted with RFC 4180
doubled-quote escaping.
"""

from __future__ import annotations

import csv
import io
from dataclasses import dataclass, field
from typing import Any, Iterable

VERSION = 1
HEADER_PREFIX = "#MUNCH/"


@dataclass
class Legends:
    """Reusable prefix table. Builds @1, @2 ... tokens by frequency."""

    prefix: str = "@"
    _counts: dict[str, int] = field(default_factory=dict)
    _order: list[str] = field(default_factory=list)
    _index: dict[str, int] = field(default_factory=dict)

    def observe(self, value: str) -> None:
        if not value:
            return
        if value not in self._counts:
            self._counts[value] = 0
            self._order.append(value)
        self._counts[value] += 1

    def finalize(self, min_uses: int = 2, min_chars_saved: int = 4) -> None:
        """Freeze the legend. Only entries that actually save bytes survive."""
        keepers: list[tuple[str, int]] = []
        for v in self._order:
            uses = self._counts[v]
            if uses < min_uses:
                continue
            # handle_len grows with index; approximate as 3 chars (@N,)
            saved = (len(v) - 3) * uses - len(v) - 3
            if saved < min_chars_saved:
                continue
            keepers.append((v, uses))
        keepers.sort(key=lambda x: -x[1] * (len(x[0]) + 1))
        self._order = [v for v, _ in keepers]
        self._index = {v: i + 1 for i, v in enumerate(self._order)}

    def encode(self, value: str) -> str:
        """Return the legend handle if known, else the raw value."""
        if value in self._index:
            return f"{self.prefix}{self._index[value]}"
        return value

    def encode_prefix(self, value: str) -> str:
        """Replace a known legend prefix at the start of value."""
        if not value:
            return value
        for v in self._order:
            if value.startswith(v):
                return f"{self.prefix}{self._index[v]}{value[len(v):]}"
        return value

    def write(self) -> str:
        if not self._order:
            return ""
        lines = [f"{self.prefix}{i + 1}={v}" for i, v in enumerate(self._order)]
        return "\n".join(lines)

    @classmethod
    def read(cls, text: str, prefix: str = "@") -> "Legends":
        leg = cls(prefix=prefix)
        for line in text.splitlines():
            if not line.startswith(prefix):
                continue
            handle, _, value = line.partition("=")
            try:
                idx = int(handle[len(prefix):])
            except ValueError:
                continue
            while len(leg._order) < idx:
                leg._order.append("")
            leg._order[idx - 1] = value
            leg._index[value] = idx
        return leg

    def decode_prefix(self, value: str) -> str:
        if not value or not value.startswith(self.prefix):
            return value
        # scan for end of handle (first non-digit after prefix)
        i = len(self.prefix)
        while i < len(value) and value[i].isdigit():
            i += 1
        if i == len(self.prefix):
            return value
        try:
            idx = int(value[len(self.prefix):i])
        except ValueError:
            return value
        if 1 <= idx <= len(self._order):
            return self._order[idx - 1] + value[i:]
        return value

    def decode(self, token: str) -> str:
        return self.decode_prefix(token)


def encode_scalar(value: Any) -> str:
    if value is None:
        return ""
    if value is True:
        return "T"
    if value is False:
        return "F"
    if isinstance(value, (int, float)):
        return repr(value) if isinstance(value, float) else str(value)
    return str(value)


def decode_scalar(raw: str, type_hint: str = "str") -> Any:
    if raw == "":
        return None
    if type_hint == "bool":
        return raw == "T"
    if type_hint == "int":
        try:
            return int(raw)
        except ValueError:
            return None
    if type_hint == "float":
        try:
            return float(raw)
        except ValueError:
            return None
    return raw


def write_header(tool: str, encoding_id: str) -> str:
    return f"{HEADER_PREFIX}{VERSION} tool={tool} enc={encoding_id}"


def parse_header(line: str) -> dict[str, str]:
    if not line.startswith(HEADER_PREFIX):
        raise ValueError(f"not a MUNCH payload: {line[:40]!r}")
    _, _, rest = line.partition(" ")
    out: dict[str, str] = {}
    for part in rest.split():
        k, _, v = part.partition("=")
        out[k] = v
    return out


def write_scalars(pairs: dict[str, Any]) -> str:
    """Single-line key=value pairs, space-separated. Values auto-quoted if needed."""
    if not pairs:
        return ""
    out: list[str] = []
    for k, v in pairs.items():
        encoded = encode_scalar(v)
        out.append(f"{k}={_quote_if_needed(encoded)}")
    return " ".join(out)


def parse_scalars(line: str) -> dict[str, str]:
    """Parse a key=value space-separated line. Respects quoted values."""
    out: dict[str, str] = {}
    for k, v in _iter_kv_tokens(line):
        out[k] = v
    return out


def _iter_kv_tokens(line: str) -> Iterable[tuple[str, str]]:
    i = 0
    n = len(line)
    while i < n:
        while i < n and line[i].isspace():
            i += 1
        if i >= n:
            break
        j = line.find("=", i)
        if j < 0:
            break
        key = line[i:j]
        i = j + 1
        if i < n and line[i] == '"':
            i += 1
            buf: list[str] = []
            while i < n:
                ch = line[i]
                if ch == '"':
                    if i + 1 < n and line[i + 1] == '"':
                        buf.append('"')
                        i += 2
                        continue
                    i += 1
                    break
                # Reverse the escapes written by _quote_if_needed so that
                # scalars containing newlines/carriage returns round-trip
                # cleanly. The `assemble` / `split_sections` layer splits on
                # `\n\n`, so raw newlines inside a quoted value would truncate
                # the scalar block mid-record (audit finding F1).
                if ch == "\\" and i + 1 < n:
                    nxt = line[i + 1]
                    if nxt == "n":
                        buf.append("\n")
                        i += 2
                        continue
                    if nxt == "r":
                        buf.append("\r")
                        i += 2
                        continue
                    if nxt == "\\":
                        buf.append("\\")
                        i += 2
                        continue
                buf.append(ch)
                i += 1
            yield key, "".join(buf)
        else:
            k2 = i
            while i < n and not line[i].isspace():
                i += 1
            yield key, line[k2:i]


def _quote_if_needed(value: str) -> str:
    if value == "":
        return '""'
    if any(c in value for c in (",", "=", " ", "\t", "\n", "\r", '"', "\\")):
        # Escape backslashes first so the follow-up substitutions remain
        # reversible, then escape the literal newlines that would otherwise
        # collide with the block separator in `assemble` (audit finding F1).
        escaped = (
            value.replace("\\", "\\\\")
                 .replace('"', '""')
                 .replace("\n", "\\n")
                 .replace("\r", "\\r")
        )
        return '"' + escaped + '"'
    return value


def write_table(tag: str, rows: Iterable[Iterable[Any]]) -> str:
    """CSV rows prefixed with a single-char tag column."""
    buf = io.StringIO()
    w = csv.writer(buf, lineterminator="\n")
    for row in rows:
        w.writerow([tag, *(encode_scalar(c) for c in row)])
    return buf.getvalue().rstrip("\n")


def read_table(text: str, tag: str) -> list[list[str]]:
    rows: list[list[str]] = []
    reader = csv.reader(io.StringIO(text))
    for row in reader:
        if row and row[0] == tag:
            rows.append(row[1:])
    return rows


def assemble(header: str, *sections: str) -> str:
    parts = [header]
    for s in sections:
        if s:
            parts.append(s)
    return "\n\n".join(parts) + "\n"


def split_sections(payload: str) -> tuple[str, list[str]]:
    """Return (header_line, [section_texts])."""
    if not payload.startswith(HEADER_PREFIX):
        raise ValueError("not a MUNCH payload")
    head, _, rest = payload.partition("\n")
    blocks = [b.strip("\n") for b in rest.split("\n\n") if b.strip()]
    return head, blocks
