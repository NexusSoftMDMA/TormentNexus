
import fs from 'fs';
import path from 'path';
import { EventEmitter } from 'events';
import { randomUUID } from 'crypto';

export interface AuditLogEntry {
    id: string;
    timestamp: number;
    type: string;
    actor?: string;
    action: string;
    outcome?: string;
    resource?: string;
    params?: unknown;
    metadata?: Record<string, unknown>;
    status?: 'SUCCESS' | 'FAILURE' | 'DENIED';
    durationMs?: number;
    agentId?: string;
    result?: unknown;
}

export interface AuditServiceOptions {
    logDir?: string;
    retentionDays?: number;
}

export class AuditService extends EventEmitter {
    private static instance: AuditService;
    private logDir: string;
    private logPath: string;
    private buffer: AuditLogEntry[] = [];
    private flushInterval: NodeJS.Timeout;
    private retentionDays: number;
    private static readonly AUTO_FLUSH_SIZE = 100;

    constructor(logDir: string = '.tormentnexus/audit', retentionDays: number = 30) {
        super();
        this.logDir = path.isAbsolute(logDir) ? logDir : path.join(process.cwd(), logDir);
        this.retentionDays = retentionDays;

        if (!fs.existsSync(this.logDir)) {
            fs.mkdirSync(this.logDir, { recursive: true });
        }
        const date = new Date().toISOString().split('T')[0];
        this.logPath = path.join(this.logDir, `audit-${date}.jsonl`);

        this.flushInterval = setInterval(() => this.flush(), 5000);
    }

    public static getInstance(options?: AuditServiceOptions): AuditService {
        if (!AuditService.instance) {
            AuditService.instance = new AuditService(options?.logDir, options?.retentionDays);
        }
        return AuditService.instance;
    }

    public dispose(): void {
        if (this.flushInterval) {
            clearInterval(this.flushInterval);
        }
    }

    // ---------- Core logging ----------

    public log(entry: {
        type: string;
        actor?: string;
        action: string;
        outcome?: string;
        resource?: string;
        params?: unknown;
        metadata?: Record<string, unknown>;
        status?: 'SUCCESS' | 'FAILURE' | 'DENIED';
        durationMs?: number;
        agentId?: string;
        result?: unknown;
    }): void {
        const fullEntry: AuditLogEntry = {
            id: `audit_${randomUUID().replace(/-/g, '').slice(0, 16)}`,
            timestamp: Date.now(),
            ...entry,
        };
        this.buffer.push(fullEntry);
        this.emit('event', fullEntry);

        if (this.buffer.length >= AuditService.AUTO_FLUSH_SIZE) {
            this.flush();
        }
    }

    // ---------- Typed helpers ----------

    public logAuth(actor: string, action: string, metadata?: Record<string, unknown>): void {
        this.log({
            type: `auth.${action}`,
            actor,
            action,
            outcome: 'success',
            metadata,
        });
    }

    public logSecretAccess(actor: string, resource: string, action: string, metadata?: Record<string, unknown>): void {
        this.log({
            type: `secret.${action}`,
            actor,
            action,
            outcome: 'success',
            resource,
            metadata,
        });
    }

    public logToolExecution(actor: string, tool: string, outcome: string, metadata?: Record<string, unknown>): void {
        this.log({
            type: outcome === 'blocked' ? 'tool.blocked' : 'tool.execute',
            actor,
            action: 'execute',
            outcome,
            resource: tool,
            metadata,
        });
    }

    // ---------- Flush ----------

    public flush(): void {
        if (this.buffer.length === 0) return;

        const chunk = this.buffer.splice(0, this.buffer.length);
        const lines = chunk.map(e => JSON.stringify(e)).join('\n') + '\n';

        try {
            fs.appendFileSync(this.logPath, lines, 'utf8');
        } catch (e) {
            console.error('[AuditService] Failed to flush logs', e);
        }
    }

    // ---------- Query ----------

    public async query(filter: {
        type?: string;
        actor?: string;
        action?: string;
        limit?: number;
        startTime?: number;
        endTime?: number;
    }): Promise<AuditLogEntry[]> {
        // Flush buffered entries so they appear in the query
        this.flush();

        if (!fs.existsSync(this.logPath)) return [];

        const content = await fs.promises.readFile(this.logPath, 'utf8');
        const lines = content.trim().split('\n');
        const logs = lines
            .map(line => {
                try { return JSON.parse(line) as AuditLogEntry; } catch { return null; }
            })
            .filter((l): l is AuditLogEntry => l !== null);

        let results = logs;
        if (filter.type) results = results.filter(l => l.type === filter.type);
        if (filter.actor) results = results.filter(l => l.actor === filter.actor);
        if (filter.action) results = results.filter(l => l.action === filter.action);
        if (filter.startTime) results = results.filter(l => l.timestamp >= filter.startTime!);
        if (filter.endTime) results = results.filter(l => l.timestamp <= filter.endTime!);

        return results.slice(-(filter.limit || 100));
    }

    /** Legacy alias for backward compatibility */
    public async queryLogs(filter?: { action?: string; limit?: number }): Promise<AuditLogEntry[]> {
        return this.query({ action: filter?.action, limit: filter?.limit ?? 100 });
    }

    // ---------- Cleanup ----------

    public async cleanup(): Promise<number> {
        const cutoff = Date.now() - this.retentionDays * 24 * 60 * 60 * 1000;
        let removed = 0;

        const files = fs.readdirSync(this.logDir).filter(f => f.startsWith('audit-') && f.endsWith('.jsonl'));
        for (const file of files) {
            const match = file.match(/audit-(\d{4}-\d{2}-\d{2})\.jsonl/);
            if (!match) continue;
            const fileDate = new Date(match[1]).getTime();
            if (fileDate < cutoff) {
                fs.unlinkSync(path.join(this.logDir, file));
                removed++;
            }
        }
        return removed;
    }
}
