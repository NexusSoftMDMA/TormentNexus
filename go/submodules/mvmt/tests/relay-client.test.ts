import { afterEach, describe, expect, it } from 'vitest';
import { createHash } from 'crypto';
import { createServer as createHttpServer } from 'http';
import net from 'net';
import { once } from 'events';
import { resolveRelayClientOptions, startRelayClient } from '../src/utils/relay-client.js';

const RELAY_ENV = ['MVMT_RELAY_URL', 'MVMT_RELAY_WORKSPACE', 'MVMT_RELAY_TOKEN'] as const;
const originalEnv = Object.fromEntries(RELAY_ENV.map((key) => [key, process.env[key]]));

afterEach(() => {
  for (const key of RELAY_ENV) {
    const value = originalEnv[key];
    if (value === undefined) delete process.env[key];
    else process.env[key] = value;
  }
});

describe('relay client options', () => {
  it('stays disabled when no relay config is present', () => {
    for (const key of RELAY_ENV) delete process.env[key];

    expect(resolveRelayClientOptions({ localPort: 4141 })).toBeUndefined();
  });

  it('requires complete relay config once any relay field is set', () => {
    for (const key of RELAY_ENV) delete process.env[key];
    process.env.MVMT_RELAY_URL = 'ws://127.0.0.1:8080/connect';

    expect(() => resolveRelayClientOptions({ localPort: 4141 }))
      .toThrow('Relay mode requires a relay URL, workspace slug, and agent token.');
  });

  it('reads saved relay tunnel config', () => {
    for (const key of RELAY_ENV) delete process.env[key];

    expect(resolveRelayClientOptions({
      localPort: 4141,
      tunnel: {
        provider: 'relay',
        relayUrl: 'ws://127.0.0.1:8080/connect',
        workspaceSlug: 'demo',
        agentToken: 'agent-secret',
      },
    })).toMatchObject({
      relayUrl: 'ws://127.0.0.1:8080/connect',
      workspaceSlug: 'demo',
      agentToken: 'agent-secret',
      localPort: 4141,
    });
  });

  it('prefers explicit CLI options over environment variables', () => {
    process.env.MVMT_RELAY_URL = 'ws://env/connect';
    process.env.MVMT_RELAY_WORKSPACE = 'env';
    process.env.MVMT_RELAY_TOKEN = 'env-token';

    expect(resolveRelayClientOptions({
      relayUrl: 'ws://cli/connect',
      workspaceSlug: 'cli',
      agentToken: 'cli-token',
      localPort: 4207,
    })).toMatchObject({
      relayUrl: 'ws://cli/connect',
      workspaceSlug: 'cli',
      agentToken: 'cli-token',
      localPort: 4207,
    });
  });

  it('reconnects when the relay socket closes', async () => {
    let connections = 0;
    let sawAuthorization = false;
    const server = net.createServer((socket) => {
      let request = '';
      let upgraded = false;
      socket.on('data', (chunk) => {
        if (upgraded) return;
        request += chunk.toString('utf8');
        if (!request.includes('\r\n\r\n')) return;
        upgraded = true;
        connections += 1;
        sawAuthorization ||= request.toLowerCase().includes('authorization: bearer agent-secret');
        const key = request.match(/^Sec-WebSocket-Key:\s*(.+)$/im)?.[1]?.trim();
        const accept = createHash('sha1')
          .update(`${key ?? ''}258EAFA5-E914-47DA-95CA-C5AB0DC85B11`)
          .digest('base64');
        socket.write([
          'HTTP/1.1 101 Switching Protocols',
          'Upgrade: websocket',
          'Connection: Upgrade',
          `Sec-WebSocket-Accept: ${accept}`,
          '',
          '',
        ].join('\r\n'));
        setTimeout(() => socket.end(), 5);
      });
    });
    server.listen(0, '127.0.0.1');
    await once(server, 'listening');
    const address = server.address();
    if (!address || typeof address === 'string') throw new Error('test server did not bind a TCP port');

    const handle = await startRelayClient({
      relayUrl: `ws://127.0.0.1:${address.port}/connect`,
      workspaceSlug: 'demo',
      agentToken: 'agent-secret',
      localPort: 4141,
      reconnectDelayMs: 10,
      pingIntervalMs: 0,
    });

    try {
      await eventually(() => connections >= 2);
      expect(sawAuthorization).toBe(true);
    } finally {
      await handle.close();
      server.close();
      await once(server, 'close');
    }
  });

  it('preserves local redirect responses and cookies for dashboard form login', async () => {
    const local = createHttpServer((req, res) => {
      if (req.method === 'POST' && req.url === '/dashboard') {
        res.statusCode = 303;
        res.setHeader('Location', 'dashboard');
        res.setHeader('Set-Cookie', 'mvmt_dashboard=session-123; Path=/; HttpOnly; SameSite=Strict; Max-Age=43200');
        res.end();
        return;
      }
      res.statusCode = 200;
      res.end('login page');
    });
    local.listen(0, '127.0.0.1');
    await once(local, 'listening');
    const localAddress = local.address();
    if (!localAddress || typeof localAddress === 'string') throw new Error('local test server did not bind a TCP port');

    let responseFrame: any;
    const responseSeen = new Promise<void>((resolve) => {
      const relay = net.createServer((socket) => {
        let request = '';
        let upgraded = false;
        let buffer = Buffer.alloc(0);
        socket.on('data', (chunk) => {
          if (!upgraded) {
            request += chunk.toString('utf8');
            if (!request.includes('\r\n\r\n')) return;
            upgraded = true;
            const key = request.match(/^Sec-WebSocket-Key:\s*(.+)$/im)?.[1]?.trim();
            const accept = createHash('sha1')
              .update(`${key ?? ''}258EAFA5-E914-47DA-95CA-C5AB0DC85B11`)
              .digest('base64');
            socket.write([
              'HTTP/1.1 101 Switching Protocols',
              'Upgrade: websocket',
              'Connection: Upgrade',
              `Sec-WebSocket-Accept: ${accept}`,
              '',
              '',
            ].join('\r\n'));
            buffer = Buffer.concat([buffer, Buffer.from(request.split('\r\n\r\n')[1] ?? '', 'binary')]);
          } else {
            buffer = Buffer.concat([buffer, chunk]);
          }
          while (true) {
            const parsed = parseTestFrame(buffer);
            if (!parsed) return;
            buffer = buffer.slice(parsed.consumed);
            if (parsed.opcode !== 0x1) continue;
            const message = JSON.parse(parsed.payload.toString('utf8'));
            if (message.type === 'hello') {
              setTimeout(() => {
                socket.write(encodeTestFrame(JSON.stringify({
                  type: 'request',
                  requestId: 'req-login',
                  method: 'POST',
                  path: '/dashboard',
                  headers: {
                    'content-type': 'application/x-www-form-urlencoded',
                    origin: 'https://mvmt-relay.fly.dev',
                  },
                  bodyBase64: Buffer.from('username=pnee&password=correct').toString('base64'),
                })));
              }, 5);
            } else if (message.type === 'response' && message.requestId === 'req-login') {
              responseFrame = message;
              resolve();
            }
          }
        });
      });
      relay.listen(0, '127.0.0.1', async () => {
        const relayAddress = relay.address();
        if (!relayAddress || typeof relayAddress === 'string') throw new Error('relay test server did not bind a TCP port');
        const handle = await startRelayClient({
          relayUrl: `ws://127.0.0.1:${relayAddress.port}/connect`,
          workspaceSlug: 'demo',
          agentToken: 'agent-secret',
          localPort: localAddress.port,
          reconnectDelayMs: 10,
          pingIntervalMs: 0,
        });
        responseSeen.finally(async () => {
          await handle.close();
          relay.close();
        });
      });
    });

    try {
      await responseSeen;
      expect(responseFrame.status).toBe(303);
      expect(responseFrame.headers.location).toBe('dashboard');
      expect(responseFrame.headers['set-cookie']).toContain('mvmt_dashboard=session-123');
    } finally {
      local.close();
      await once(local, 'close');
    }
  });

  it('reassembles fragmented relay requests before forwarding to local mvmt', async () => {
    let uploadedBody = '';
    const local = createHttpServer((req, res) => {
      if (req.method !== 'PUT' || req.url !== '/upload') {
        res.statusCode = 404;
        res.end('not found');
        return;
      }
      const chunks: Buffer[] = [];
      req.on('data', (chunk) => chunks.push(Buffer.from(chunk)));
      req.on('end', () => {
        uploadedBody = Buffer.concat(chunks).toString('utf8');
        res.statusCode = 201;
        res.end('stored');
      });
    });
    local.listen(0, '127.0.0.1');
    await once(local, 'listening');
    const localAddress = local.address();
    if (!localAddress || typeof localAddress === 'string') throw new Error('local test server did not bind a TCP port');

    let responseFrame: any;
    const responseSeen = new Promise<void>((resolve) => {
      const relay = net.createServer((socket) => {
        let request = '';
        let upgraded = false;
        let buffer = Buffer.alloc(0);
        socket.on('data', (chunk) => {
          if (!upgraded) {
            request += chunk.toString('utf8');
            if (!request.includes('\r\n\r\n')) return;
            upgraded = true;
            const key = request.match(/^Sec-WebSocket-Key:\s*(.+)$/im)?.[1]?.trim();
            const accept = createHash('sha1')
              .update(`${key ?? ''}258EAFA5-E914-47DA-95CA-C5AB0DC85B11`)
              .digest('base64');
            socket.write([
              'HTTP/1.1 101 Switching Protocols',
              'Upgrade: websocket',
              'Connection: Upgrade',
              `Sec-WebSocket-Accept: ${accept}`,
              '',
              '',
            ].join('\r\n'));
            buffer = Buffer.concat([buffer, Buffer.from(request.split('\r\n\r\n')[1] ?? '', 'binary')]);
          } else {
            buffer = Buffer.concat([buffer, chunk]);
          }
          while (true) {
            const parsed = parseTestFrame(buffer);
            if (!parsed) return;
            buffer = buffer.slice(parsed.consumed);
            if (parsed.opcode !== 0x1) continue;
            const message = JSON.parse(parsed.payload.toString('utf8'));
            if (message.type === 'hello') {
              const requestFrame = JSON.stringify({
                type: 'request',
                requestId: 'req-upload',
                method: 'PUT',
                path: '/upload',
                headers: { 'content-type': 'text/plain' },
                bodyBase64: Buffer.from('hello from relay').toString('base64'),
              });
              const splitAt = Math.floor(requestFrame.length / 2);
              socket.write(Buffer.concat([
                encodeTestFrame(requestFrame.slice(0, splitAt), { fin: false, opcode: 0x1 }),
                encodeTestFrame(requestFrame.slice(splitAt), { fin: true, opcode: 0x0 }),
              ]));
            } else if (message.type === 'response' && message.requestId === 'req-upload') {
              responseFrame = message;
              resolve();
            }
          }
        });
      });
      relay.listen(0, '127.0.0.1', async () => {
        const relayAddress = relay.address();
        if (!relayAddress || typeof relayAddress === 'string') throw new Error('relay test server did not bind a TCP port');
        const handle = await startRelayClient({
          relayUrl: `ws://127.0.0.1:${relayAddress.port}/connect`,
          workspaceSlug: 'demo',
          agentToken: 'agent-secret',
          localPort: localAddress.port,
          reconnectDelayMs: 10,
          pingIntervalMs: 0,
        });
        responseSeen.finally(async () => {
          await handle.close();
          relay.close();
        });
      });
    });

    try {
      await responseSeen;
      expect(uploadedBody).toBe('hello from relay');
      expect(responseFrame.status).toBe(201);
      expect(Buffer.from(responseFrame.bodyBase64, 'base64').toString('utf8')).toBe('stored');
    } finally {
      local.close();
      await once(local, 'close');
    }
  });
});

async function eventually(condition: () => boolean): Promise<void> {
  const deadline = Date.now() + 1_000;
  while (Date.now() < deadline) {
    if (condition()) return;
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
  expect(condition()).toBe(true);
}

function encodeTestFrame(
  message: string,
  options: { fin?: boolean; opcode?: number } = {},
): Buffer {
  const payload = Buffer.from(message, 'utf8');
  const firstByte = (options.fin ?? true ? 0x80 : 0) | (options.opcode ?? 0x1);
  if (payload.length < 126) return Buffer.concat([Buffer.from([firstByte, payload.length]), payload]);
  const header = Buffer.alloc(4);
  header[0] = firstByte;
  header[1] = 126;
  header.writeUInt16BE(payload.length, 2);
  return Buffer.concat([header, payload]);
}

function parseTestFrame(buffer: Buffer): { opcode: number; payload: Buffer; consumed: number } | undefined {
  if (buffer.length < 2) return undefined;
  const opcode = buffer[0]! & 0x0f;
  let length = buffer[1]! & 0x7f;
  let offset = 2;
  if (length === 126) {
    if (buffer.length < 4) return undefined;
    length = buffer.readUInt16BE(2);
    offset = 4;
  }
  const masked = Boolean(buffer[1]! & 0x80);
  const maskOffset = masked ? offset : -1;
  if (masked) offset += 4;
  if (buffer.length < offset + length) return undefined;
  const payload = Buffer.from(buffer.slice(offset, offset + length));
  if (masked) {
    const mask = buffer.slice(maskOffset, maskOffset + 4);
    for (let i = 0; i < payload.length; i += 1) payload[i] = payload[i]! ^ mask[i % 4]!;
  }
  return { opcode, payload, consumed: offset + length };
}
