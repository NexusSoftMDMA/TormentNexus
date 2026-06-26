import fs from 'fs';
import os from 'os';
import path from 'path';

export const AUDIT_LOG_PATH = path.join(os.homedir(), '.mvmt', 'audit.log');

export interface AuditEntry {
  ts: string;
  event?: 'token.add' | 'token.edit' | 'token.rotate' | 'token.remove' | 'token.use';
  connectorId: string;
  tool: string;
  clientId?: string;
  name?: string;
  scope?: string;
  client?: string;
  expires?: string | null;
  result?: 'success' | 'error';
  argKeys: string[];
  argPreview: string;
  isError: boolean;
  deniedReason?: string;
  durationMs: number;
}

// HTTP request audit entry. Structurally identical to the server's
// HttpRequestLogEntry; redeclared here to avoid a dependency from utils
// back into the server module. Persisted to the same audit.log file as
// tool-call entries, tagged with `type: 'http'` so consumers can split.
export interface HttpAuditEntry {
  ts: string;
  kind: string;
  method: string;
  path: string;
  status: number;
  detail?: string;
  clientId?: string;
  ip?: string;
}

export interface AuditLogger {
  record(entry: AuditEntry): void;
  recordHttp(entry: HttpAuditEntry): void;
}

export function createAuditLogger(logPath: string = AUDIT_LOG_PATH): AuditLogger {
  fs.mkdirSync(path.dirname(logPath), { recursive: true });

  if (!fs.existsSync(logPath)) {
    fs.writeFileSync(logPath, '', { mode: 0o600 });
  }
  if (process.platform !== 'win32') {
    try {
      fs.chmodSync(logPath, 0o600);
    } catch {
      // Best-effort permission hardening.
    }
  }

  function append(line: string): void {
    try {
      fs.appendFileSync(logPath, `${line}\n`);
    } catch {
      // Never let audit logging break a tool call or request.
    }
  }

  return {
    record(entry: AuditEntry): void {
      append(JSON.stringify(entry));
    },
    recordHttp(entry: HttpAuditEntry): void {
      append(JSON.stringify({ type: 'http', ...entry }));
    },
  };
}

export function summarizeArgs(args: Record<string, unknown>): { argKeys: string[]; argPreview: string } {
  const argKeys = Object.keys(args);
  let preview: string;
  try {
    preview = JSON.stringify(args);
  } catch {
    preview = '[unserializable]';
  }
  if (preview.length > 512) preview = `${preview.slice(0, 509)}...`;
  return { argKeys, argPreview: preview };
}
