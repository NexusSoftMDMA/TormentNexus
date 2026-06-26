import { AgentAdapter, AgentConfig } from '../AgentAdapter.js';
import { GeminiAgent } from '../../agents/GeminiAgent.js';

export class GeminiAdapter extends AgentAdapter {
    private agent: GeminiAgent | null = null;

    constructor() {
        const config: AgentConfig = {
            name: 'Gemini',
            command: 'internal', // Virtual process
            args: [],
        };
        super(config);
    }

    async start(): Promise<void> {
        if (!this.agent) {
            const mcp = global.mcpServerInstance;
            if (mcp) {
                // Initialize internal agent
                // Assuming MCPServer creates PromptRegistry or we create it here.
                // Ideally PromptRegistry should be on MCPServer. 
                // We will patch MCPServer in next step to add PromptRegistry.
                // Accessing dependencies via global or passing them down.

                // Creating fresh registry if needed, but optimally share it.
                // We will refactor MCPServer to expose promptRegistry later.
                const registry = mcp.promptRegistry;

                this.agent = new GeminiAgent(mcp.llmService, registry);

                this.agent.on('output', (data) => this.emit('output', data));
                this.agent.on('error', (err) => this.emit('error', err.message));

                await this.agent.start();
            }
        }
    }

    async stop(): Promise<void> {
        this.agent = null;
    }

    async send(input: string): Promise<void> {
        if (this.agent) {
            await this.agent.send(input);
        } else {
            throw new Error("Gemini Agent not started");
        }
    }

    isActive(): boolean {
        return this.agent !== null && this.agent.isActive();
    }
}