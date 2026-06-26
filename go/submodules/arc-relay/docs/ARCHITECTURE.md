# Arc Relay - Architecture

## Overview

Arc Relay is a lightweight management system for deploying, proxying, and sharing MCP (Model Context Protocol) servers. It consolidates multiple MCP servers behind a single gateway with simple authentication and RBAC, making it easy to expose MCP capabilities to AI tools like Claude Desktop, Claude Code, and others.

**Goals:** Simpler alternative to [microsoft/mcp-gateway](https://github.com/microsoft/mcp-gateway) - no Kubernetes, no Azure dependencies, no .NET. Just a single Go binary + Docker.

---

## Architecture

```
+-------------------------------------------------------------+
|                        Arc Relay                            |
|                                                             |
|  +----------+  +--------------+  +-----------------------+  |
|  | Web UI   |  | Admin API    |  | MCP Proxy Layer       |  |
|  | (HTML    |  | (REST)       |  |                       |  |
|  | templates)|  |              |  | /mcp/pfsense-prod ------>| Docker: pfsense-mcp (stdio)
|  |          |  |              |  | /mcp/pfsense-dev  ------>| Docker: pfsense-mcp (stdio)
|  |          |  |              |  | /mcp/uptime-kuma  ------>| Docker: uptime-kuma-mcp (HTTP)
|  |          |  |              |  | /mcp/homeassistant ----->| Remote: HA MCP server
|  |          |  |              |  | /mcp/sentry ------------>| Remote: mcp.sentry.dev (OAuth)
|  +----------+  +--------------+  +-----------------------+  |
|                                                             |
|  +--------------+  +--------------+  +------------------+   |
|  | Auth Layer   |  | Server Mgr   |  | Docker Client    |   |
|  | (API keys,   |  | (lifecycle,  |  | (container mgmt) |   |
|  |  sessions)   |  |  health)     |  |                  |   |
|  +--------------+  +--------------+  +------------------+   |
|                                                             |
|  +------------------------------------------------------+   |
|  | SQLite Database                                      |   |
|  | (servers, users, API keys, RBAC, logs)               |   |
|  +------------------------------------------------------+   |
+-------------------------------------------------------------+
```

### Key Architectural Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | **Go** | Single binary, excellent networking/proxy, good subprocess mgmt, low memory |
| Stdio server mgmt | **Docker containers** | Isolation, reproducibility, health checks, no dependency conflicts |
| Frontend | **Server-rendered HTML** (Go templates) | No build step, no JS framework, fast to build |
| Routing | **Path-based** on single port | Simpler networking, easy reverse proxy, single TLS cert |
| Database | **SQLite** | Zero-config, embedded, sufficient for this scale |
| Config | **TOML** for app config, DB for server/user state | TOML is readable, DB for dynamic state |

---

## MCP Server Types

Arc Relay supports three server types, each with different lifecycle management:

### 1. Stdio (Docker-wrapped)

The server runs as a subprocess inside a Docker container managed by Arc Relay. Arc Relay communicates with it over stdin/stdout via `docker exec` or by running a bridge process inside the container.

**Lifecycle:** Arc Relay builds/pulls the image, starts the container, and manages the stdio bridge. The bridge translates between Streamable HTTP (exposed to clients) and stdio (to the server process).

**Examples:** pfSense MCP Server (Python, stdio)

**Config inputs:**
- Docker image (or Dockerfile/repo URL to build from)
- Environment variables (key-value pairs, stored encrypted)
- Optional: custom command, working directory

### 2. HTTP (Docker or external)

The server exposes an HTTP endpoint (Streamable HTTP or legacy SSE). It may run in a Docker container managed by Arc Relay, or be an external service.

**Lifecycle:** For Docker-managed, Arc Relay starts the container and proxies to its HTTP port. For external, Arc Relay just proxies.

**Examples:** Uptime Kuma MCP Server (Python, SSE/HTTP on port 8000)

**Config inputs:**
- Docker image + port mapping, OR external URL
- Environment variables
- Health check endpoint (optional)

### 3. Remote

The server is hosted externally. Arc Relay acts as a pure proxy, forwarding MCP protocol messages.

**Lifecycle:** No lifecycle management. Arc Relay stores connection details and credentials, proxies requests.

**Examples:**
- Home Assistant MCP (HA add-on, private URL with embedded auth token)
- Sentry MCP (OAuth flow via `mcp.sentry.dev`)

**Config inputs:**
- Remote URL (may contain embedded auth, as with HA's private URL scheme)
- Auth type: none, private URL (auth in URL), bearer token, API key header, or OAuth
- OAuth config: client ID, auth URL, token URL, scopes (for OAuth servers)

---

## Core Features

### F1: Add/Manage MCP Server

**Web UI flow:**
1. User clicks "Add Server"
2. Selects type: Stdio (Docker), HTTP (Docker), HTTP (External), Remote
3. Fills in config form:
   - Name (slug, used in URL path)
   - Display name
   - Type-specific fields (image, env vars, URL, auth)
4. System validates config, pulls/builds image if needed
5. System starts server (if managed) and runs MCP `initialize` to verify connectivity
6. Server appears in dashboard

**API:**
```
POST   /api/servers           - Create server
GET    /api/servers           - List servers
GET    /api/servers/:id       - Get server detail
PUT    /api/servers/:id       - Update server
DELETE /api/servers/:id       - Delete server
POST   /api/servers/:id/start - Start managed server
POST   /api/servers/:id/stop  - Stop managed server
```

### F2: List Servers & Enumerate Endpoints

Once a server is running and connected, Arc Relay calls `tools/list`, `resources/list`, and `prompts/list` on the server and caches the results.

**Web UI:** Dashboard shows all servers with status, and expandable sections showing their tools, resources, and prompts.

**RBAC per endpoint:**
- Each tool/resource/prompt can be allowed or denied per user/role
- Default: all endpoints allowed for all users
- Admin can toggle individual endpoint access per user

### F3: Proxy MCP Servers

Each server is exposed at `/mcp/{server-name}`. Arc Relay implements Streamable HTTP transport (the current MCP standard) on the client-facing side, regardless of the backend server's transport.

**Proxy flow:**
```
AI Client (Claude, etc.)
    |
    |  Streamable HTTP (POST/GET with SSE)
    |  + Auth header (Bearer token)
    v
Arc Relay (/mcp/{server-name})
    |
    +-> Auth check (validate token, check user permissions)
    +-> RBAC filter (strip disallowed tools/resources from responses)
    |
    +-> [stdio server] --> Docker container stdin/stdout
    +-> [http server]  --> HTTP proxy to container or external URL
    +-> [remote server] -> HTTP proxy to remote URL (with stored credentials)
```

**Session management:** Arc Relay manages sessions per client connection. For stdio backends, it maintains the subprocess session. For HTTP/remote backends, it forwards session IDs.

**Key proxy behaviors:**
- `initialize` requests: forwarded, response cached for endpoint enumeration
- `tools/list`: response filtered by RBAC before returning to client
- `tools/call`: checked against RBAC, forwarded if allowed, error if denied
- Same pattern for `resources/*` and `prompts/*`

### F4: User Management

Simple user system with bcrypt password hashing and SHA-256 API key storage.

**Auth methods:**
- **Web UI:** Session cookie (login with username/password)
- **MCP proxy endpoints:** Bearer token (API key)

**Web UI pages:**
- User list (admin only)
- Create/edit user
- API key management (generate, revoke, list)

### F5: Logging

All MCP requests proxied through Arc Relay are logged, including method, endpoint name, duration, status, and error (if any). For stdio servers, stderr output is captured and stored separately.

### F6: Connection Config Generation

Generate ready-to-paste config snippets for connecting to Arc Relay servers.

**Claude Desktop (`claude_desktop_config.json`):**
```json
{
  "mcpServers": {
    "pfsense-prod": {
      "url": "http://arc-relay.local:8080/mcp/pfsense-prod",
      "headers": {
        "Authorization": "Bearer <your-api-key>"
      }
    }
  }
}
```

**Claude Code:**
```bash
claude mcp add --transport http pfsense-prod http://arc-relay.local:8080/mcp/pfsense-prod --header "Authorization: Bearer <your-api-key>"
```

### F7: Access Logs & Analytics

Web UI dashboards show request counts per server, per endpoint, per user; error rates; response-time percentiles; and timeline charts.

---

## Stdio-to-HTTP Bridge Design

This is the most complex piece. Arc Relay needs to translate between Streamable HTTP (what clients connect with) and stdio (what the server subprocess speaks).

### Approach: Bridge Process per Connection

```
Client --HTTP--> Arc Relay --stdin/stdout--> Docker Container
                     |                              |
                     |  Manages session              |  Runs MCP server
                     |  Translates HTTP<->stdio      |  (e.g., pfsense-mcp)
                     |  Buffers JSON-RPC messages    |
```

**Implementation:**
1. Docker container runs the MCP server process (e.g., `python -m pfsense_mcp`)
2. Arc Relay attaches to the container's stdin/stdout via Docker API (`ContainerAttach`)
3. For each client session:
   - Client POSTs a JSON-RPC message to `/mcp/{server-name}`
   - Arc Relay writes the message + newline to the container's stdin
   - Arc Relay reads the response from stdout (newline-delimited JSON-RPC)
   - Response is returned to client as JSON or SSE stream

**Concurrency consideration:** Stdio is inherently single-session. Options:
- **One container per client session** - simplest, most isolated, but resource-heavy
- **Multiplexed access with request queuing** - single container, serialize requests, match responses by JSON-RPC id. Most efficient.

Arc Relay uses the multiplexed approach with per-server request queuing.

---

## Example Server Configurations

### Stdio (Docker)

```json
{
  "name": "pfsense-prod",
  "display_name": "pfSense - Production Firewall",
  "server_type": "stdio",
  "config": {
    "image": "ghcr.io/your-org/pfsense-mcp-server:latest",
    "command": ["python", "-m", "pfsense_mcp"],
    "env": {
      "PFSENSE_URL": "https://pfsense.example.com",
      "PFSENSE_API_KEY": "encrypted:...",
      "AUTH_METHOD": "api_key"
    }
  }
}
```

### HTTP (Docker)

```json
{
  "name": "uptime-kuma",
  "display_name": "Uptime Kuma Monitoring",
  "server_type": "http",
  "config": {
    "image": "ghcr.io/example/uptime-kuma-mcp-server:latest",
    "port": 8000,
    "env": {
      "KUMA_URL": "http://uptime-kuma.example.com:3001",
      "KUMA_USERNAME": "admin",
      "KUMA_PASSWORD": "encrypted:..."
    },
    "health_check": "/health"
  }
}
```

### Remote (private URL auth)

```json
{
  "name": "homeassistant",
  "display_name": "Home Assistant",
  "server_type": "remote",
  "config": {
    "url": "http://homeassistant.example.com:8123/api/mcp/xxxxxxxxxxxxxxxx",
    "auth": {
      "type": "private_url"
    }
  }
}
```

> **Note:** Home Assistant's MCP add-on exposes a Streamable HTTP endpoint with a private URL that embeds the auth token in the path. No separate token header is needed.

### Remote (OAuth)

```json
{
  "name": "sentry",
  "display_name": "Sentry Error Tracking",
  "server_type": "remote",
  "config": {
    "url": "https://mcp.sentry.dev/mcp",
    "auth": {
      "type": "oauth",
      "auth_url": "https://sentry.io/oauth/authorize/",
      "token_url": "https://sentry.io/oauth/token/",
      "client_id": "...",
      "client_secret": "encrypted:...",
      "scopes": ["org:read", "project:read", "event:read"]
    }
  }
}
```

---

For the on-disk project layout, see [AGENTS.md](../AGENTS.md). For server configuration and environment variables, see [`config.example.toml`](../config.example.toml) and the README.
