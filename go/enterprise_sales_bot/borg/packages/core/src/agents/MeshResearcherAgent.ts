import { SpecializedAgent } from '../mesh/SpecializedAgent.js';
import { MCPServer } from '../MCPServer.js';

export class MeshResearcherAgent extends SpecializedAgent {
    private mcpServer: MCPServer;

    constructor(mcpServer: MCPServer) {
        // Expose 'researcher' capability to the mesh
        super('MeshResearcher', ['researcher']);
        this.mcpServer = mcpServer;
    }

    public async handleTask(offer: any): Promise<any> {
        console.log(`[MeshResearcher] 📚 Received research task from Swarm: "${offer.task}"`);

        try {
            // Dispatch to the internal ResearcherAgent implementation
            if (!this.mcpServer.researcherAgent) {
                // Fallback if not specifically instantiated, use deep research
                const res = await this.mcpServer.deepResearchService.recursiveResearch(offer.task, 2, 3);
                return { result: res };
            }

            const result = await this.mcpServer.researcherAgent.handleTask({
                task: offer.task,
                worktreePath: offer.worktreePath,
                options: { depth: 2, breadth: 3 }
            });

            return { result };
        } catch (e: any) {
            console.error(`[MeshResearcher] Error: ${e.message}`);
            throw e;
        }
    }
}
