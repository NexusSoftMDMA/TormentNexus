import { SpecializedAgent } from '../mesh/SpecializedAgent.js';
import { MCPServer } from '../MCPServer.js';

export class MeshCoderAgent extends SpecializedAgent {
    private mcpServer: MCPServer;

    constructor(mcpServer: MCPServer) {
        // Expose 'coder' capability to the mesh
        super('MeshCoder', ['coder']);
        this.mcpServer = mcpServer;
    }

    public async handleTask(offer: any): Promise<any> {
        console.log(`[MeshCoder] 💻 Received coding task from Swarm: "${offer.task}"`);

        try {
            if (!this.mcpServer.coderAgent) {
                throw new Error("Local CoderAgent is not initialized on this node.");
            }

            // Dispatch to the internal CoderAgent implementation
            const result = await this.mcpServer.coderAgent.handleTask({
                task: offer.task,
                worktreePath: offer.worktreePath
            });

            // The Mesh expects the final payload to be primarily a string `result` field
            // But we can return objects which get stringified in executeTask
            return { result: JSON.stringify(result, null, 2) };
        } catch (e: any) {
            console.error(`[MeshCoder] Error: ${e.message}`);
            throw e;
        }
    }
}
