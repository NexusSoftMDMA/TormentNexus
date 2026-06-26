import { Buffer } from 'buffer';
import { createHash, randomBytes } from 'crypto';
import net from 'net';
import tls from 'tls';
import { once } from 'events';
import { TunnelConfig } from '../config/schema.js';
import { Logger } from './logger.js';

type RelayRequestFrame = {
  type: 'request';
  requestId: string;
  method: string;
  path: string;
  headers?: Record<string, string>;
  bodyBase64?: string;
};

type RelayResponseFrame = {
  type: 'response';
  requestId: string;
  status: number;
  headers: Record<string, string>;
  bodyBase64: string;
};

type RelayErrorFrame = {
  type: 'error';
  requestId: string;
  status: number;
  message: string;
};

export type RelayClientOptions = {
  relayUrl: string;
  workspaceSlug: string;
  agentToken: string;
  localPort: number;
  logger?: Logger;
  reconnectDelayMs?: number;
  pingIntervalMs?: number;
};

export type RelayClientHandle = {
  close(): Promise<void>;
  isConnected(): boolean;
};

export function resolveRelayClientOptions(input: {
  relayUrl?: string;
  workspaceSlug?: string;
  agentToken?: string;
  localPort: number;
  logger?: Logger;
  tunnel?: TunnelConfig;
}): RelayClientOptions | undefined {
  const relayTunnel = input.tunnel?.provider === 'relay' ? input.tunnel : undefined;
  const relayUrl = input.relayUrl ?? relayTunnel?.relayUrl ?? process.env.MVMT_RELAY_URL;
  const workspaceSlug = input.workspaceSlug ?? relayTunnel?.workspaceSlug ?? process.env.MVMT_RELAY_WORKSPACE;
  const agentToken = input.agentToken ?? relayTunnel?.agentToken ?? process.env.MVMT_RELAY_TOKEN;
  if (!relayUrl && !workspaceSlug && !agentToken) return undefined;
  if (!relayUrl || !workspaceSlug || !agentToken) {
    throw new Error('Relay mode requires a relay URL, workspace slug, and agent token.');
  }
  return { relayUrl, workspaceSlug, agentToken, localPort: input.localPort, logger: input.logger };
}

export async function startRelayClient(options: RelayClientOptions): Promise<RelayClientHandle> {
  let closed = false;
  let current: RelayWebSocket | undefined;
  let reconnectTimer: NodeJS.Timeout | undefined;
  let pingTimer: NodeJS.Timeout | undefined;
  const reconnectDelayMs = options.reconnectDelayMs ?? 2_000;
  const pingIntervalMs = options.pingIntervalMs ?? 25_000;

  const hello = () => JSON.stringify({
    type: 'hello',
    workspaceId: `local-${options.workspaceSlug}`,
    workspaceSlug: options.workspaceSlug,
    protocolVersion: 'v1',
  });

  const clearReconnectTimer = () => {
    if (!reconnectTimer) return;
    clearTimeout(reconnectTimer);
    reconnectTimer = undefined;
  };

  const clearPingTimer = () => {
    if (!pingTimer) return;
    clearInterval(pingTimer);
    pingTimer = undefined;
  };

  const scheduleReconnect = () => {
    if (closed || reconnectTimer) return;
    reconnectTimer = setTimeout(() => {
      reconnectTimer = undefined;
      void connect(false).catch((err) => {
        if (closed) return;
        options.logger?.warn(`Relay reconnect failed: ${err instanceof Error ? err.message : 'unknown error'}`);
        scheduleReconnect();
      });
    }, reconnectDelayMs);
    reconnectTimer.unref();
  };

  const connect = async (initial: boolean) => {
    const ws = await connectWebSocket(options.relayUrl, {
      Authorization: `Bearer ${options.agentToken}`,
    });
    if (closed) {
      ws.close();
      return;
    }

    current = ws;
    ws.send(hello());

    ws.onMessage((message) => {
      void handleRelayMessage(ws, message, options).catch((err) => {
        options.logger?.warn(`Relay request failed: ${err instanceof Error ? err.message : 'unknown error'}`);
      });
    });
    ws.onClose(() => {
      if (current === ws) current = undefined;
      clearPingTimer();
      if (!closed) {
        options.logger?.warn(`Relay connection closed; reconnecting in ${Math.round(reconnectDelayMs / 1000)}s.`);
        scheduleReconnect();
      }
    });

    if (pingIntervalMs > 0) {
      clearPingTimer();
      pingTimer = setInterval(() => {
        try {
          ws.send(JSON.stringify({ type: 'ping' }));
        } catch {
          // The close handler owns reconnect scheduling.
        }
      }, pingIntervalMs);
      pingTimer.unref();
    }

    options.logger?.info(`${initial ? 'Relay connected' : 'Relay reconnected'}: ${options.workspaceSlug}`);
  };

  await connect(true);

  return {
    close: async () => {
      closed = true;
      clearReconnectTimer();
      clearPingTimer();
      current?.close();
      current = undefined;
    },
    isConnected: () => Boolean(current),
  };
}

async function handleRelayMessage(ws: RelayWebSocket, raw: string, options: RelayClientOptions): Promise<void> {
  const frame = parseFrame(raw);
  if (!frame || frame.type !== 'request') return;

  try {
    const response = await forwardToLocalMvmt(frame, options.localPort);
    ws.send(JSON.stringify(response));
  } catch (err) {
    const errorFrame: RelayErrorFrame = {
      type: 'error',
      requestId: frame.requestId,
      status: 502,
      message: err instanceof Error ? err.message : 'local mvmt request failed',
    };
    ws.send(JSON.stringify(errorFrame));
  }
}

type RelayWebSocket = {
  send(message: string): void;
  close(): void;
  onMessage(listener: (message: string) => void): void;
  onClose(listener: () => void): void;
};

async function connectWebSocket(rawUrl: string, headers: Record<string, string>): Promise<RelayWebSocket> {
  const url = new URL(rawUrl);
  const secure = url.protocol === 'wss:';
  if (!secure && url.protocol !== 'ws:') throw new Error('Relay URL must use ws:// or wss://.');

  const port = Number(url.port || (secure ? 443 : 80));
  const socket = secure
    ? tls.connect({ host: url.hostname, port, servername: url.hostname })
    : net.connect({ host: url.hostname, port });
  await once(socket, 'connect');
  socket.setKeepAlive(true, 30_000);

  const key = randomBytes(16).toString('base64');
  const path = `${url.pathname || '/'}${url.search}`;
  const requestHeaders = [
    `GET ${path} HTTP/1.1`,
    `Host: ${url.host}`,
    'Upgrade: websocket',
    'Connection: Upgrade',
    `Sec-WebSocket-Key: ${key}`,
    'Sec-WebSocket-Version: 13',
    ...Object.entries(headers).map(([name, value]) => `${name}: ${value}`),
    '',
    '',
  ].join('\r\n');
  socket.write(requestHeaders);

  let handshake = Buffer.alloc(0);
  while (true) {
    const [chunk] = await once(socket, 'data') as [Buffer];
    handshake = Buffer.concat([handshake, chunk]);
    const end = handshake.indexOf('\r\n\r\n');
    if (end === -1) continue;
    const head = handshake.slice(0, end).toString('utf8');
    const rest = handshake.slice(end + 4);
    verifyHandshake(head, key);
    return createRelayWebSocket(socket, rest);
  }
}

function verifyHandshake(head: string, key: string): void {
  const lines = head.split('\r\n');
  if (!/^HTTP\/1\.[01] 101\b/.test(lines[0] ?? '')) {
    throw new Error(`Relay WebSocket upgrade failed: ${lines[0] ?? 'no response'}`);
  }
  const headers = new Map<string, string>();
  for (const line of lines.slice(1)) {
    const separator = line.indexOf(':');
    if (separator > 0) headers.set(line.slice(0, separator).toLowerCase(), line.slice(separator + 1).trim());
  }
  const expected = createHash('sha1')
    .update(key + '258EAFA5-E914-47DA-95CA-C5AB0DC85B11')
    .digest('base64');
  if (headers.get('sec-websocket-accept') !== expected) {
    throw new Error('Relay WebSocket upgrade returned an invalid accept key.');
  }
}

function createRelayWebSocket(socket: net.Socket | tls.TLSSocket, initial: Buffer): RelayWebSocket {
  let buffer = initial;
  const messageListeners = new Set<(message: string) => void>();
  const closeListeners = new Set<() => void>();
  let closeNotified = false;
  let textFragments: Buffer[] | undefined;

  const notifyClose = () => {
    if (closeNotified) return;
    closeNotified = true;
    for (const listener of closeListeners) listener();
  };

  const parseAvailableFrames = () => {
    while (true) {
      const parsed = parseWebSocketFrame(buffer);
      if (!parsed) return;
      buffer = buffer.slice(parsed.consumed);
      if (parsed.opcode === 0x1) {
        if (parsed.fin) {
          const message = parsed.payload.toString('utf8');
          for (const listener of messageListeners) listener(message);
        } else {
          textFragments = [parsed.payload];
        }
      } else if (parsed.opcode === 0x0) {
        if (!textFragments) continue;
        textFragments.push(parsed.payload);
        if (parsed.fin) {
          const message = Buffer.concat(textFragments).toString('utf8');
          textFragments = undefined;
          for (const listener of messageListeners) listener(message);
        }
      } else if (parsed.opcode === 0x8) {
        socket.end();
        notifyClose();
        return;
      } else if (parsed.opcode === 0x9) {
        socket.write(encodeWebSocketFrame(parsed.payload, 0xA));
      }
    }
  };

  socket.on('data', (chunk) => {
    buffer = Buffer.concat([buffer, chunk]);
    parseAvailableFrames();
  });
  socket.on('close', () => {
    notifyClose();
  });
  socket.on('error', () => {
    notifyClose();
  });
  parseAvailableFrames();

  return {
    send: (message) => socket.write(encodeWebSocketFrame(Buffer.from(message, 'utf8'), 0x1)),
    close: () => {
      if (!socket.destroyed) {
        socket.write(encodeWebSocketFrame(Buffer.alloc(0), 0x8));
        socket.end();
      }
    },
    onMessage: (listener) => {
      messageListeners.add(listener);
    },
    onClose: (listener) => {
      closeListeners.add(listener);
    },
  };
}

function parseWebSocketFrame(buffer: Buffer): { fin: boolean; opcode: number; payload: Buffer; consumed: number } | undefined {
  if (buffer.length < 2) return undefined;
  const fin = Boolean(buffer[0]! & 0x80);
  const opcode = buffer[0]! & 0x0f;
  const masked = Boolean(buffer[1]! & 0x80);
  let length = buffer[1]! & 0x7f;
  let offset = 2;
  if (length === 126) {
    if (buffer.length < offset + 2) return undefined;
    length = buffer.readUInt16BE(offset);
    offset += 2;
  } else if (length === 127) {
    if (buffer.length < offset + 8) return undefined;
    const high = buffer.readUInt32BE(offset);
    const low = buffer.readUInt32BE(offset + 4);
    if (high !== 0) throw new Error('Relay frame is too large.');
    length = low;
    offset += 8;
  }
  const maskOffset = masked ? offset : -1;
  if (masked) offset += 4;
  if (buffer.length < offset + length) return undefined;
  const payload = Buffer.from(buffer.slice(offset, offset + length));
  if (masked) {
    const mask = buffer.slice(maskOffset, maskOffset + 4);
    for (let i = 0; i < payload.length; i += 1) payload[i] = payload[i]! ^ mask[i % 4]!;
  }
  return { fin, opcode, payload, consumed: offset + length };
}

function encodeWebSocketFrame(payload: Buffer, opcode: number): Buffer {
  const mask = randomBytes(4);
  const headerLength = payload.length < 126 ? 2 : payload.length <= 0xffff ? 4 : 10;
  const frame = Buffer.alloc(headerLength + 4 + payload.length);
  frame[0] = 0x80 | opcode;
  if (payload.length < 126) {
    frame[1] = 0x80 | payload.length;
    mask.copy(frame, 2);
    maskPayload(payload, mask).copy(frame, 6);
  } else if (payload.length <= 0xffff) {
    frame[1] = 0x80 | 126;
    frame.writeUInt16BE(payload.length, 2);
    mask.copy(frame, 4);
    maskPayload(payload, mask).copy(frame, 8);
  } else {
    frame[1] = 0x80 | 127;
    frame.writeUInt32BE(0, 2);
    frame.writeUInt32BE(payload.length, 6);
    mask.copy(frame, 10);
    maskPayload(payload, mask).copy(frame, 14);
  }
  return frame;
}

function maskPayload(payload: Buffer, mask: Buffer): Buffer {
  const out = Buffer.from(payload);
  for (let i = 0; i < out.length; i += 1) out[i] = out[i]! ^ mask[i % 4]!;
  return out;
}

async function forwardToLocalMvmt(frame: RelayRequestFrame, port: number): Promise<RelayResponseFrame> {
  const url = new URL(frame.path || '/', `http://127.0.0.1:${port}`);
  const headers = new Headers();
  for (const [key, value] of Object.entries(frame.headers ?? {})) {
    if (key.toLowerCase() === 'host' || key.toLowerCase() === 'content-length') continue;
    headers.set(key, value);
  }

  const method = frame.method || 'GET';
  const body = method === 'GET' || method === 'HEAD'
    ? undefined
    : Buffer.from(frame.bodyBase64 ?? '', 'base64');
  const response = await fetch(url, { method, headers, body, redirect: 'manual' });
  const bytes = Buffer.from(await response.arrayBuffer());
  return {
    type: 'response',
    requestId: frame.requestId,
    status: response.status,
    headers: responseHeaders(response.headers),
    bodyBase64: bytes.toString('base64'),
  };
}

function responseHeaders(headers: Headers): Record<string, string> {
  const out: Record<string, string> = {};
  headers.forEach((value, key) => {
    if (key.toLowerCase() === 'content-length') return;
    out[key] = value;
  });
  return out;
}

function parseFrame(raw: unknown): RelayRequestFrame | undefined {
  const text = typeof raw === 'string'
    ? raw
    : raw instanceof Buffer
      ? raw.toString('utf8')
      : String(raw);
  try {
    const parsed = JSON.parse(text);
    if (!parsed || typeof parsed !== 'object') return undefined;
    if (parsed.type !== 'request' || typeof parsed.requestId !== 'string') return undefined;
    return parsed as RelayRequestFrame;
  } catch {
    return undefined;
  }
}
