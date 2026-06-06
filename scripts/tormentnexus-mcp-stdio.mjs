/**
 * TormentNexus MCP Stdio Bridge
 * 
 * Bridges pi-agent's stdio MCP protocol to the TormentNexus Go HTTP server.
 * Uses line-delimited JSON-RPC over stdio.
 */

const PORT = process.env.TORMENTNEXUS_PORT || '4300';
const HOST = process.env.TORMENTNEXUS_HOST || '127.0.0.1';
const BASE = `http://${HOST}:${PORT}`;

function send(msg) {
  process.stdout.write(JSON.stringify(msg) + '\n');
}

async function apiGet(path) {
  try {
    const resp = await fetch(`${BASE}${path}`);
    if (resp.ok) return await resp.json();
    return null;
  } catch { return null; }
}

async function apiPost(path, body) {
  try {
    const resp = await fetch(`${BASE}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (resp.ok) return await resp.json();
    return null;
  } catch { return null; }
}

const handlers = {
  'initialize': (id, _params) => ({
    jsonrpc: '2.0', id,
    result: {
      protocolVersion: '2024-11-05',
      capabilities: { tools: { listChanged: false } },
      serverInfo: { name: 'tormentnexus', version: '1.0.0-alpha.125' },
    }
  }),

  'tools/list': async (id, _params) => {
    const data = await apiGet('/api/native/tools/list');
    const tools = (data?.tools || []).map(t => ({
      name: t.name,
      description: t.description || `Go-native: ${t.name}`,
      inputSchema: { type: 'object', properties: { args: { type: 'object', description: 'Tool arguments' } } },
    }));
    return { jsonrpc: '2.0', id, result: { tools } };
  },

  'tools/call': async (id, params) => {
    const name = params?.name;
    const args = params?.arguments || {};
    const data = await apiPost('/api/mcp/tools/call', { name, args });
    return {
      jsonrpc: '2.0', id,
      result: {
        content: [{ type: 'text', text: JSON.stringify(data || { error: 'tool call failed' }, null, 2) }],
      }
    };
  },

  'resources/list': (id) => ({ jsonrpc: '2.0', id, result: { resources: [] } }),
  'prompts/list': (id) => ({ jsonrpc: '2.0', id, result: { prompts: [] } }),
  'ping': (id) => ({ jsonrpc: '2.0', id, result: {} }),
};

async function handleLine(line) {
  if (!line.trim()) return;
  let msg;
  try { msg = JSON.parse(line.trim()); } catch { return; }
  
  const { id, method, params } = msg;
  if (!method) return;
  
  // Notifications don't need responses
  if (method.startsWith('notifications/')) return;
  
  const handler = handlers[method];
  if (!handler) {
    send({ jsonrpc: '2.0', id, error: { code: -32601, message: `Method not found: ${method}` } });
    return;
  }
  
  try {
    const result = typeof handler === 'function' 
      ? await handler(id, params) 
      : handler(id, params);
    if (result) send(result);
  } catch (err) {
    send({ jsonrpc: '2.0', id, error: { code: -32603, message: err.message } });
  }
}

// Line-delimited JSON-RPC stdio
let buffer = '';
process.stdin.on('data', (chunk) => {
  buffer += chunk.toString();
  const lines = buffer.split('\n');
  buffer = lines.pop();
  for (const line of lines) handleLine(line);
});

process.stdin.on('end', () => {
  if (buffer.trim()) handleLine(buffer);
  setTimeout(() => process.exit(0), 100);
});

// Keep alive
setInterval(() => {}, 60000);
