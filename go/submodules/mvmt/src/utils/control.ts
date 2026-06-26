import crypto from 'crypto';
import fs from 'fs';
import net from 'net';
import os from 'os';
import path from 'path';
import { getConfigPath } from '../config/loader.js';

const MAX_CONTROL_BUFFER_BYTES = 64 * 1024;

export class ControlUnavailableError extends Error {}
export class ControlAuthError extends Error {}

export interface JsonControlConnection {
  send(message: unknown): void;
  close(): void;
  onClose(listener: () => void): void;
}

export interface JsonControlServer {
  close(): Promise<void>;
}

export function getControlSocketPath(configPath = getConfigPath()): string {
  const hash = crypto.createHash('sha256').update(configPath).digest('hex').slice(0, 12);
  if (process.platform === 'win32') {
    return `\\\\.\\pipe\\mvmt-${hash}`;
  }
  return path.join(os.homedir(), '.mvmt', `control-${hash}.sock`);
}

export interface StartControlServerOptions {
  verifyToken: (token: unknown) => boolean;
}

export async function startJsonControlServer(
  socketPath: string,
  onMessage: (message: any, connection: JsonControlConnection) => void | Promise<void>,
  options: StartControlServerOptions,
): Promise<JsonControlServer> {
  if (process.platform !== 'win32') {
    fs.mkdirSync(path.dirname(socketPath), { recursive: true });
    try {
      fs.unlinkSync(socketPath);
    } catch (err) {
      if ((err as NodeJS.ErrnoException).code !== 'ENOENT') throw err;
    }
  }

  const server = net.createServer((socket) => {
    socket.setEncoding('utf-8');
    let buffer = '';
    const closeListeners = new Set<() => void>();

    const connection: JsonControlConnection = {
      send(message) {
        socket.write(`${JSON.stringify(message)}\n`);
      },
      close() {
        socket.end();
      },
      onClose(listener) {
        closeListeners.add(listener);
      },
    };

    socket.on('data', async (chunk) => {
      buffer += chunk;
      if (buffer.length > MAX_CONTROL_BUFFER_BYTES) {
        connection.send({ ok: false, error: 'Control request exceeds maximum size' });
        connection.close();
        buffer = '';
        return;
      }
      const lines = buffer.split('\n');
      buffer = lines.pop() ?? '';

      for (const line of lines.map((value) => value.trim()).filter(Boolean)) {
        let message: any;
        try {
          message = JSON.parse(line);
        } catch (err) {
          connection.send({
            ok: false,
            error: err instanceof Error ? err.message : 'Invalid control request',
          });
          connection.close();
          return;
        }

        if (!options.verifyToken(message?.token)) {
          connection.send({ ok: false, error: 'Invalid or missing control token' });
          connection.close();
          return;
        }

        try {
          // Strip auth token before dispatching so handlers can't accidentally
          // echo it in logs or responses.
          const { token: _token, ...payload } = message;
          await onMessage(payload, connection);
        } catch (err) {
          connection.send({
            ok: false,
            error: err instanceof Error ? err.message : 'Control request failed',
          });
          connection.close();
        }
      }
    });

    socket.on('close', () => {
      for (const listener of closeListeners) listener();
    });
  });

  await new Promise<void>((resolve, reject) => {
    const onError = (err: Error) => reject(err);
    server.once('error', onError);
    server.listen(socketPath, () => {
      server.off('error', onError);
      try {
        hardenControlSocket(socketPath);
      } catch (err) {
        const error = err instanceof Error ? err : new Error('Failed to harden control socket permissions');
        server.close(() => {
          removeControlSocket(socketPath);
          reject(error);
        });
        return;
      }
      resolve();
    });
  });

  return {
    async close() {
      await new Promise<void>((resolve, reject) => {
        server.close((err) => (err ? reject(err) : resolve()));
      });
      if (process.platform !== 'win32') {
        removeControlSocket(socketPath);
      }
    },
  };
}

function hardenControlSocket(socketPath: string): void {
  if (process.platform === 'win32') return;
  fs.chmodSync(socketPath, 0o600);
}

function removeControlSocket(socketPath: string): void {
  try {
    fs.unlinkSync(socketPath);
  } catch (err) {
    if ((err as NodeJS.ErrnoException).code !== 'ENOENT') throw err;
  }
}

export async function sendJsonControlRequest<T>(
  socketPath: string,
  message: Record<string, unknown>,
  token: string,
): Promise<T> {
  const response = await new Promise<any>((resolve, reject) => {
    const socket = net.createConnection(socketPath);
    socket.setEncoding('utf-8');
    let buffer = '';
    let settled = false;

    socket.once('error', (err: NodeJS.ErrnoException) => {
      if (settled) return;
      settled = true;
      if (err.code === 'ENOENT' || err.code === 'ECONNREFUSED') {
        reject(new ControlUnavailableError('mvmt is not running for this config.'));
        return;
      }
      reject(err);
    });

    socket.on('data', (chunk) => {
      if (settled) return;
      buffer += chunk;
      if (buffer.length > MAX_CONTROL_BUFFER_BYTES) {
        settled = true;
        socket.destroy();
        reject(new Error('Control response exceeds maximum size'));
        return;
      }
      const lines = buffer.split('\n');
      buffer = lines.pop() ?? '';
      for (const line of lines.map((value) => value.trim()).filter(Boolean)) {
        settled = true;
        try {
          resolve(JSON.parse(line));
        } catch (err) {
          reject(err instanceof Error ? err : new Error('Invalid control response'));
        }
        socket.end();
        return;
      }
    });

    socket.on('connect', () => {
      socket.write(`${JSON.stringify({ ...message, token })}\n`);
    });
  });

  if (!response?.ok) {
    const error = response?.error ?? 'Control request failed';
    if (typeof error === 'string' && error.toLowerCase().includes('control token')) {
      throw new ControlAuthError(error);
    }
    throw new Error(error);
  }
  return response.result as T;
}

export async function streamJsonControl(
  socketPath: string,
  message: Record<string, unknown>,
  onMessage: (message: any) => void,
  token: string,
): Promise<() => void> {
  const socket = net.createConnection(socketPath);
  socket.setEncoding('utf-8');
  let buffer = '';
  let startupSettled = false;
  let streamEnded = false;

  const emitEnd = () => {
    if (streamEnded) return;
    streamEnded = true;
    onMessage({ kind: 'end' });
  };

  await new Promise<void>((resolve, reject) => {
    socket.once('error', (err: NodeJS.ErrnoException) => {
      if (err.code === 'ENOENT' || err.code === 'ECONNREFUSED') {
        reject(new ControlUnavailableError('mvmt is not running for this config.'));
        return;
      }
      reject(err);
    });

    socket.once('close', () => {
      if (!startupSettled) {
        reject(new Error('Control stream closed before initialization'));
        return;
      }
      emitEnd();
    });

    socket.on('data', (chunk) => {
      buffer += chunk;
      if (buffer.length > MAX_CONTROL_BUFFER_BYTES) {
        socket.destroy();
        if (!startupSettled) {
          reject(new Error('Control stream exceeds maximum size'));
          return;
        }
        onMessage({ kind: 'error', error: 'Control stream exceeds maximum size' });
        return;
      }
      const lines = buffer.split('\n');
      buffer = lines.pop() ?? '';
      for (const line of lines.map((value) => value.trim()).filter(Boolean)) {
        try {
          const parsed = JSON.parse(line);
          if (parsed?.ok === false) {
            const error = parsed?.error ?? 'Control stream failed';
            const failure =
              typeof error === 'string' && error.toLowerCase().includes('control token')
                ? new ControlAuthError(error)
                : new Error(String(error));
            socket.end();
            if (!startupSettled) {
              reject(failure);
              return;
            }
            onMessage({ kind: 'error', error: failure.message });
            return;
          }
          if (!startupSettled) {
            startupSettled = true;
            resolve();
          }
          onMessage(parsed);
        } catch {
          socket.end();
          if (!startupSettled) {
            reject(new Error('Invalid control stream message'));
            return;
          }
          onMessage({ kind: 'error', error: 'Invalid control stream message' });
        }
      }
    });

    socket.on('connect', () => {
      socket.write(`${JSON.stringify({ ...message, token })}\n`);
    });
  });

  return () => {
    socket.end();
  };
}
