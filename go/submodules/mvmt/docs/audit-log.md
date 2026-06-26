# Audit Log

Every tool call routed through mvmt is appended to:

```text
~/.mvmt/audit.log
```

The file is JSONL and is created with mode `600` on non-Windows systems.

> [!WARNING]
> `argPreview` can include truncated argument values. Do not place secrets in tool arguments unless you are comfortable with those values appearing in the local audit log.

## Format

Example mount-tool call:

```json
{
  "ts": "2026-04-14T14:23:01.442Z",
  "connectorId": "mvmt",
  "tool": "search",
  "clientId": "chatgpt",
  "argKeys": ["query", "limit"],
  "argPreview": "{\"query\":\"meeting notes\",\"limit\":10}",
  "isError": false,
  "durationMs": 34
}
```

| Field | Description |
| --- | --- |
| `ts` | ISO 8601 timestamp. |
| `connectorId` | Runtime surface that handled the call. Mount tools use `mvmt`. |
| `tool` | Tool name used by the MCP client, such as `search` or `read`. |
| `clientId` | Present when HTTP auth resolved to a configured or legacy client identity. |
| `argKeys` | Argument key names, without values. |
| `argPreview` | Truncated JSON of the arguments, max 512 characters. Can contain values. |
| `isError` | `true` if the tool returned an error or the call threw. |
| `deniedReason` | Present when mvmt denied a tool call before data access. |
| `durationMs` | Time from call start to result, in milliseconds. |

## Querying

View recent activity:

```bash
tail -20 ~/.mvmt/audit.log | jq .
```

Filter denied calls:

```bash
jq 'select(.deniedReason != null)' ~/.mvmt/audit.log
```

Count tool calls:

```bash
jq -r '.tool' ~/.mvmt/audit.log | sort | uniq -c | sort -rn
```

Filter by client:

```bash
jq 'select(.clientId == "codex")' ~/.mvmt/audit.log
```

## HTTP/OAuth Logs

OAuth discovery, registration, token exchange, and MCP session setup failures happen before a tool call exists. Those events do not go into `audit.log`.

Use interactive mode for sanitized request logs:

```bash
mvmt serve -i
> logs on
```

## Rotation

mvmt does not rotate the audit log automatically.

Archive manually:

```bash
mv ~/.mvmt/audit.log ~/.mvmt/audit.log.bak
```

mvmt creates a new log on the next tool call.
