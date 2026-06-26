
import { execSync } from 'child_process';
import os from 'os';
import fs from 'fs';
import path from 'path';

import { getBridgeHealthUrl } from '../bridge/bridgePort.js';

export interface HealthStatus {
    component: string;
    status: 'online' | 'offline' | 'warning';
    details: string;
}

export class DiagnosticTools {
    public static async checkHealth(): Promise<HealthStatus[]> {
        const results: HealthStatus[] = [];

        // 1. Core API
        try {
            const res = await fetch(getBridgeHealthUrl(), { signal: AbortSignal.timeout(3000) });
            if (res.ok) {
                results.push({ component: 'Core API', status: 'online', details: 'Healthy' });
            } else {
                results.push({ component: 'Core API', status: 'warning', details: `Status ${res.status}` });
            }
        } catch (e) {
            results.push({ component: 'Core API', status: 'offline', details: 'Connection refused' });
        }

        // 2. Web Dashboard
        try {
            const res = await fetch('http://localhost:3000', { signal: AbortSignal.timeout(3000) });
            if (res.ok) {
                results.push({ component: 'Web UI', status: 'online', details: 'Dashboard active' });
            } else {
                results.push({ component: 'Web UI', status: 'warning', details: 'Dashboard offline' });
            }
        } catch (e) {
            results.push({ component: 'Web UI', status: 'offline', details: 'Dashboard not running' });
        }

        // 3. VS Code Bridge
        // Placeholder check - usually checked via WebSocket clients count in MCPServer
        results.push({
            component: 'VS Code Bridge',
            status: 'warning',
            details: 'Connection checked only on-demand via WS'
        });

        // 4. System Stats
        const load = os.loadavg();
        const freeMem = os.freemem() / 1024 / 1024 / 1024;
        results.push({
            component: 'Host System',
            status: 'online',
            details: `Load: ${load[0].toFixed(2)}, Free RAM: ${freeMem.toFixed(2)}GB`
        });

        return results;
    }

    public static getDiagnosticsMarkup(status: HealthStatus[]): string {
        let markup = "## 🩺 System Health Report\n\n";
        markup += "| Component | Status | Details |\n";
        markup += "|-----------|--------|---------|\n";
        for (const s of status) {
            const icon = s.status === 'online' ? '✅' : s.status === 'warning' ? '⚠️' : '❌';
            markup += `| ${s.component} | ${icon} ${s.status.toUpperCase()} | ${s.details} |\n`;
        }
        return markup;
    }
}
