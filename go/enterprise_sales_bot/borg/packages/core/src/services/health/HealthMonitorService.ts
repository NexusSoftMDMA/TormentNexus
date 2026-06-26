/**
 * HealthMonitorService.ts
 * 
 * Production telemetry and auto-restart capabilities.
 * Hooks into Node.js process events to track uptime, memory, and handle graceful shutdown.
 */

import os from 'os';
import { EventEmitter } from 'events';

export class HealthMonitorService extends EventEmitter {
    private startTime: number;
    private maxMemoryMb: number;
    private interval: NodeJS.Timeout | null = null;

    constructor(maxMemoryMb: number = 2048) {
        super();
        this.startTime = Date.now();
        this.maxMemoryMb = maxMemoryMb;
    }

    start(checkIntervalMs: number = 30000) {
        if (this.interval) return;

        this.interval = setInterval(() => this.checkHealth(), checkIntervalMs);

        // Handle uncaught errors gracefully
        process.on('uncaughtException', (err) => {
            console.error('CRITICAL: Uncaught Exception', err);
            this.emit('crash', { reason: 'uncaughtException', error: err.message });
            this.gracefulShutdown(1);
        });

        process.on('unhandledRejection', (reason) => {
            console.error('CRITICAL: Unhandled Rejection', reason);
            this.emit('crash', { reason: 'unhandledRejection', error: String(reason) });
        });
    }

    stop() {
        if (this.interval) {
            clearInterval(this.interval);
            this.interval = null;
        }
    }

    getStats() {
        const mem = process.memoryUsage();
        return {
            uptimeSeconds: Math.floor((Date.now() - this.startTime) / 1000),
            memory: {
                rss: Math.floor(mem.rss / 1024 / 1024) + ' MB',
                heapTotal: Math.floor(mem.heapTotal / 1024 / 1024) + ' MB',
                heapUsed: Math.floor(mem.heapUsed / 1024 / 1024) + ' MB'
            },
            cpu: process.cpuUsage(),
            loadavg: os.loadavg(),
            status: 'healthy'
        };
    }

    private checkHealth() {
        const mem = process.memoryUsage();
        const heapUsedMb = mem.heapUsed / 1024 / 1024;

        this.emit('heartbeat', this.getStats());

        // Auto-restart threshold check
        if (heapUsedMb > this.maxMemoryMb) {
            console.warn(`[HEALTH] Memory exceeded ${this.maxMemoryMb}MB. Current: ${heapUsedMb.toFixed(2)}MB. Triggering restart.`);
            this.emit('crash', { reason: 'OOM_PREVENTION' });
            this.gracefulShutdown(1);
        }
    }

    private gracefulShutdown(code: number) {
        console.log("Initiating graceful shutdown...");
        // In a real production setup, we'd close db connections, finish requests, then exit.
        setTimeout(() => process.exit(code), 2000); // Force exit after 2s
    }
}
