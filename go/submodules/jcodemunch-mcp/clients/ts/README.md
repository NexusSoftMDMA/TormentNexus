# MUNCH TypeScript decoder

Reference TypeScript decoder for jcodemunch's MUNCH compact output format.
Companion to the canonical Python decoder at
`jcodemunch_mcp.encoding.decoder.decode`.

## Install

No npm package (yet). Copy `decoder.ts` into your project — it has zero
dependencies, ~200 lines, works in any modern JS runtime.

## Usage

```ts
import { decode } from "./decoder";

// Works on any tool-response string — falls through to JSON.parse for
// payloads that aren't MUNCH.
const obj = decode(payload);
```

## What it does

- Parses the `#MUNCH/1` header.
- Reads the optional legend section (`@N=prefix`) into a handle→literal map.
- Parses the scalars section (RFC-4180-quoted key=value pairs).
- Reads table sections (single-character-tagged CSV rows).
- Reconstructs the original response using the embedded `__tables` schema
  and optional `__stypes` scalar-type map.
- Handles `_meta.X` flattening, `parent.child` dotted nested dicts, and
  `__json.X` JSON-blob scalars.

Both tier-1 encoder payloads (`fr1`, `dg1`, etc.) and generic-fallback
payloads (`gen1`) round-trip through this single decoder.

## Spec

Full on-wire format documented at
[`SPEC_MUNCH.md`](../../SPEC_MUNCH.md) in the repo root.

## Fallback

If your client can't decode MUNCH, request `format="json"` on the tool call
and receive a standard JSON dict.
