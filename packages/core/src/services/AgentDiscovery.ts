import { spawnAsync } from '../utils/exec.js';
import * as fs from 'fs/promises';
import * as path from 'path';

export interface DiscoveredAgent {
    name: string;
    executable: string;
    version: string;
    status: 'available' | 'missing';
}

export interface AgentCapability {
    id: string;
    name: string;
    version?: string;
    path: string;
    type: 'cli' | 'service' | 'mcp';
    features: string[];
}

export class AgentDiscovery {
    private discoveredAgents: Map<string, AgentCapability> = new Map();

    async discover(): Promise<AgentCapability[]> {
        console.log('[Core:Discovery] Scanning for local agents...');

        const scanTasks = [
            this.scanForClaudeCode(),
            this.scanForElectronOrchestrator(),
            this.scanForCloudOrchestrator(),
            this.scanForStandardTools()
        ];

        const results = await Promise.all(scanTasks);
        results.flat().forEach(agent => {
            if (agent) this.discoveredAgents.set(agent.id, agent);
        });

        return Array.from(this.discoveredAgents.values());
    }

    private async scanForClaudeCode(): Promise<AgentCapability | null> {
        try {
            const isWin = process.platform === 'win32';
            const cmd = isWin ? 'where.exe' : 'which';
            const result = await spawnAsync(cmd, ['claude']);
            const claudePath = result.stdout.split('\n')[0].trim();

            if (claudePath) {
                return {
                    id: 'claude-code',
                    name: 'Claude Code',
                    path: claudePath,
                    type: 'cli',
                    features: ['coding', 'terminal', 'mcp']
                };
            }
        } catch {
            // Claude not found
        }
        return null;
    }

    private async scanForElectronOrchestrator(): Promise<AgentCapability | null> {
        const electronOrchestratorPath = path.resolve(process.cwd(), 'apps/maestro');
        try {
            await fs.access(electronOrchestratorPath);
            return {
                id: 'electron-orchestrator',
                name: 'electron-orchestrator',
                path: electronOrchestratorPath,
                type: 'service',
                features: ['orchestration', 'ui', 'multi-agent']
            };
        } catch {
            return null;
        }
    }

    private async scanForCloudOrchestrator(): Promise<AgentCapability | null> {
        const candidatePaths = [
            path.resolve(process.cwd(), 'apps/cloud-orchestrator'),
            path.resolve(process.cwd(), 'jules-autopilot'),
        ];
        for (const candidatePath of candidatePaths) {
            try {
                await fs.access(candidatePath);
                return {
                    id: 'cloud-orchestrator',
                    name: 'cloud-orchestrator',
                    path: candidatePath,
                    type: 'service',
                    features: ['autopilot', 'debate', 'risk-scoring']
                };
            } catch {
            }
        }
        return null;
    }

    private async scanForStandardTools(): Promise<AgentCapability[]> {
        return [];
    }

    getAgent(id: string): AgentCapability | undefined {
        return this.discoveredAgents.get(id);
    }
}
