
import fs from 'fs';
import path from 'path';

export interface AuditLogEntry {
    timestamp: number;
    action: string;
    params: unknown;
    level: string;
}

export class AuditService {
    private logPath: string;
    private buffer: AuditLogEntry[] = [];
    private flushInterval: NodeJS.Timeout;

    constructor(logDir: string = '.tormentnexus/audit') {
        const absoluteLogDir = path.isAbsolute(logDir) ? logDir : path.join(process.cwd(), logDir);
        if (!fs.existsSync(absoluteLogDir)) {
            fs.mkdirSync(absoluteLogDir, { recursive: true });
        }
        // Rotate logs by day
        const date = new Date().toISOString().split('T')[0];
        this.logPath = path.join(absoluteLogDir, `audit-${date}.jsonl`);

        // Auto-flush every 5 seconds
        this.flushInterval = setInterval(() => this.flush(), 5000);
    }

    public log(action: string, params: unknown, level: string = 'INFO') {
        const entry: AuditLogEntry = {
            timestamp: Date.now(),
            action,
            params,
            level
        };
        this.buffer.push(entry);
    }

    public async flush() {
        if (this.buffer.length === 0) return;

        const chunk = this.buffer.splice(0, this.buffer.length);
        const lines = chunk.map(e => JSON.stringify(e)).join('\n') + '\n';

        try {
            await fs.promises.appendFile(this.logPath, lines, 'utf8');
        } catch (e) {
            console.error("[AuditService] Failed to flush logs", e);
        }
    }

    public async queryLogs(limit: number = 100): Promise<AuditLogEntry[]> {
        if (!fs.existsSync(this.logPath)) return [];

        try {
            const content = await fs.promises.readFile(this.logPath, 'utf8');
            const lines = content.trim().split('\n');
            return lines
                .map(line => {
                    try { return JSON.parse(line); } catch { return null; }
                })
                .filter(l => l !== null)
                .slice(-limit)
                .reverse(); // Newest first
        } catch (e) {
            console.error("[AuditService] Failed to query logs", e);
            return [];
        }
    }
    public async getLogs(limit: number = 50): Promise<AuditLogEntry[]> {
        return this.queryLogs(limit);
    }

    public async query(filter: { level?: string, agentId?: string, action?: string, limit?: number }): Promise<AuditLogEntry[]> {
        const logs = await this.queryLogs(filter.limit || 100);
        return logs.filter(log => {
            if (filter.level && log.level !== filter.level) return false;
            if (filter.action && log.action !== filter.action) return false;
            return true;
        });
    }
}
