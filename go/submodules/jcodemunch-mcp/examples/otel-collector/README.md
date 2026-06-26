# OTel Collector → jcodemunch-mcp live ingest

Drop-in example for streaming OTel spans, SQL query logs, and application stack traces to a running jcodemunch-mcp instance via the Phase 6 HTTP live-ingest endpoint.

## Two-key turn

The endpoint is **off by default**. Both must be set on the jcodemunch-mcp host:

```bash
export JCODEMUNCH_HTTP_TOKEN=<long-random-string>
export JCODEMUNCH_RUNTIME_INGEST_ENABLED=1
jcodemunch-mcp serve --transport sse --host 0.0.0.0 --port 7331
```

Without the bearer token the HTTP transport refuses every request from a non-loopback host. Without `RUNTIME_INGEST_ENABLED` the runtime routes return 503 even with a valid token. Both flags are deliberate — the runtime endpoint *writes* to the index and we want operators to make that decision twice.

## Wire it up

`jcm-exporter.yaml` ships three `otlphttp/*` exporter blocks — one per source. Copy the ones you need into your existing collector config and add them to the relevant pipelines under `service.pipelines.*.exporters`.

Set `JCM_TOKEN` in the collector's environment so the bearer auth header gets populated at runtime.

The repo identifier is per-exporter via the `X-JCM-Repo` header; if your collector handles multiple repos in one pipeline, instead pass `?repo=owner/name` via the endpoint URL on each exporter (one exporter per repo).

## Verify it's flowing

After spans start arriving, run:

```python
get_redaction_log(repo="myorg/myrepo")
get_runtime_coverage(repo="myorg/myrepo")
```

The first proves the redaction chokepoint fired; the second proves data landed in the index. If `runtime_redaction_log` rows aren't growing within seconds, the request never made it through — check token, ingest-enabled flag, and (most commonly) the X-JCM-Repo header.

## Endpoint reference

```
POST /runtime/otel    body: OTLP/JSON spans (file-exporter format works as-is)
POST /runtime/sql     body: pg_stat_statements CSV  (?fmt=csv)
                            generic SQL JSON-Lines  (?fmt=jsonl)
                            auto-detect             (?fmt=auto, default)
POST /runtime/stack   body: plain-text app log     (?fmt=plain)
                            JSON-Lines records     (?fmt=jsonl)
                            auto-detect            (?fmt=auto, default)
```

Headers: `Authorization: Bearer <JCM_TOKEN>` (required), `X-JCM-Repo: owner/name` (or `?repo=` query), `Content-Encoding: gzip` (optional; honoured under the body-size cap).

Default body cap: 5 MB after decompression. Override with `JCODEMUNCH_RUNTIME_INGEST_MAX_BODY_BYTES`.

## Why not a single endpoint?

Three URLs is more honest about what's actually different between the sources — the parser, the schema rows that get touched, and the `_meta` envelope. One endpoint with a `?source=` flag would just hide the same complexity in a query string. Different routes also let you put separate rate-limits or alerting on each source if you need to.
