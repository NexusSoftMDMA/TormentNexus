
import fs from 'fs/promises';
import path from 'path';
import os from 'os';

export class Logger {
    private logPath: string;

    constructor() {
        const homeDir = os.homedir();
        const logDir = path.join(homeDir, '.tormentnexus', 'logs');
        this.logPath = path.join(logDir, 'supervisor.log');

        // Ensure log directory exists
        fs.mkdir(logDir, { recursive: true }).catch(err => console.error("Failed to create log dir", err));
    }

    async log(level: 'INFO' | 'ERROR' | 'WARN', message: string, data?: any) {
        const timestamp = new Date().toISOString();
        const logEntry = `[${timestamp}] [${level}] ${message} ${data ? JSON.stringify(data) : ''}\n`;

        // Write to Console (Stdio)
        if (level === 'ERROR') {
            console.error(logEntry.trim());
        } else {
            console.error(logEntry.trim()); // MCP uses stdout for protocol, so logic uses stderr for logs
        }

        // Write to File
        try {
            await fs.appendFile(this.logPath, logEntry, 'utf-8');
        } catch (err) {
            // Fail silently if logging fails to avoid crashing the supervisor
        }
    }

    info(message: string, data?: any) {
        this.log('INFO', message, data);
    }

    error(message: string, data?: any) {
        this.log('ERROR', message, data);
    }

    warn(message: string, data?: any) {
        this.log('WARN', message, data);
    }
}

export const logger = new Logger();
