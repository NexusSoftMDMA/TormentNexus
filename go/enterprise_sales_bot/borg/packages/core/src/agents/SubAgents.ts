
import { AgentStatus, BaseAgent } from './BaseAgent.js';
import type { MCPServer } from '../MCPServer.js';

export class ResearchAgent extends BaseAgent {
    private server: MCPServer;

    constructor(task: string, server: MCPServer) {
        super('research', task);
        this.server = server;
    }

    async run(): Promise<void> {
        this.status = AgentStatus.RUNNING;
        this.log('Starting research task...');

        try {
            this.log('Dispatching to researcher agent...');

            const agentResult = this.server.researcherAgent
                ? await this.server.researcherAgent.handleTask({ task: this.task, options: { depth: 2, breadth: 3 } })
                : await this.server.deepResearchService.recursiveResearch(this.task, 2, 3);

            const output = typeof agentResult === 'string'
                ? agentResult
                : JSON.stringify(agentResult, null, 2);

            let telemetry = '';
            if (typeof agentResult === 'object' && agentResult !== null && 'modelMetadata' in agentResult) {
                const meta = (agentResult as any).modelMetadata;
                if (meta?.provider && meta?.modelId) {
                    telemetry = `\n[Telemetry: Executed via ${meta.provider}/${meta.modelId}]`;
                }
            }

            this.complete(`Research complete for: ${this.task}\n\n${output}${telemetry}`);
        } catch (e: any) {
            this.fail(e.message);
        }
    }
}

export class CodeAgent extends BaseAgent {
    private server: MCPServer;

    constructor(task: string, server: MCPServer) {
        super('code', task);
        this.server = server;
    }

    async run(): Promise<void> {
        this.status = AgentStatus.RUNNING;
        this.log('Starting coding task...');

        try {
            this.log('Dispatching to coder agent...');

            if (!this.server.coderAgent) {
                throw new Error('Coder agent is not initialized');
            }

            const result = await this.server.coderAgent.handleTask({ task: this.task });
            const changed = Array.isArray(result?.filesChanged) ? result.filesChanged.join(', ') : 'none';
            const reasoning = result?.reasoning ? `\nReasoning: ${result.reasoning}` : '';

            let telemetry = '';
            if (result?.modelMetadata?.provider && result?.modelMetadata?.modelId) {
                telemetry = `\n[Telemetry: Executed via ${result.modelMetadata.provider}/${result.modelMetadata.modelId}]`;
            }

            this.complete(`Coding task complete for: ${this.task}\n\nFiles changed: ${changed}${reasoning}${telemetry}`);
        } catch (e: any) {
            this.fail(e.message);
        }
    }
}
