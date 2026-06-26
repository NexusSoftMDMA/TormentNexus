# Agent hints — MUNCH compact output

When an agent calls a jcodemunch-mcp tool with `format="auto"` (the default)
or `format="compact"`, the response may arrive as a MUNCH-encoded string
instead of a JSON dict. This file contains drop-in prompt snippets so the
agent can read those payloads directly, without a client-side decoder.

---

## Short snippet (for general agent prompts)

Paste this into the agent's system prompt:

```
Some jcodemunch-mcp tool responses arrive as MUNCH-encoded strings, not JSON.
A MUNCH payload begins with `#MUNCH/1 tool=... enc=...` followed by
blank-line-separated sections:

  1. Optional legend lines `@N=prefix` — N is an integer handle; any value
     starting with `@N` is that prefix + the remainder.
  2. One scalar line of `key=value` pairs (space separated; values with
     spaces/commas/equals are double-quoted, doubled-quote escaped).
     Reserved keys:
       - `__tables`: comma-separated list, each `tag:key:col1|col2|...:t1|t2|...`
       - `__stypes`: pipe-separated `name:type` map for non-string top-level scalars
       - `_meta.<X>`: flattened meta fields (re-nest into `_meta.{X}`)
       - `<parent>.<child>`: flattened nested dict (re-nest into `parent.{child}`)
       - `__json.<X>`: JSON-encoded blob — parse as JSON, store at key `X`
  3. Zero or more table sections of CSV rows where column 0 is a 1-char tag
     matching an entry in `__tables`. Rows with that tag belong to that table;
     remaining columns are fields in the order declared by the schema.

Types in table schemas are `int` / `float` / `bool` / `str`.
`T`=true, `F`=false, empty field=null.

If you cannot parse MUNCH, re-issue the call with `format="json"`.
```

---

## Full spec

The canonical on-wire format lives in [SPEC_MUNCH.md](SPEC_MUNCH.md). The
Python reference decoder is at `jcodemunch_mcp.encoding.decoder.decode`;
the TypeScript reference decoder is at `clients/ts/decoder.ts`.

---

## Escape hatch

Any client (or agent) that cannot decode MUNCH can:

1. Pass `format="json"` on the tool call — the server will skip encoding
   and return a plain JSON dict for that call.
2. Set the env var `JCODEMUNCH_DEFAULT_FORMAT=json` on the server to
   disable compact encoding globally.

---

## Minimal worked example

Encoded payload:

```
#MUNCH/1 tool=find_references enc=fr1

@1=src/models/

repo=acme identifier=get_user reference_count=3 __tables=r:references:file|line|kind:str|int|str _meta.timing_ms=3.1 _meta.truncated=F

r,@1a.py,10,call
r,@1a.py,22,call
r,@1b.py,5,ref
```

Decoded:

```json
{
  "repo": "acme",
  "identifier": "get_user",
  "reference_count": "3",
  "references": [
    {"file": "src/models/a.py", "line": 10, "kind": "call"},
    {"file": "src/models/a.py", "line": 22, "kind": "call"},
    {"file": "src/models/b.py", "line":  5, "kind": "ref"}
  ],
  "_meta": {"timing_ms": 3.1, "truncated": false}
}
```

Note `reference_count` is `"3"` (string) because it has no entry in
`__stypes`; top-level scalars decode as strings unless typed. Table column
types (`int` for `line`) coerce their values during rehydration.
