import { SpecializedAgent } from '../mesh/SpecializedAgent.js';
import { DeepResearchService } from '../services/DeepResearchService.js';
import { A2AMessage, A2AMessageType, IA2AClient, A2ATask } from "@tormentnexus/adk";
import { a2aBroker } from "@tormentnexus/agents";

/**
 * ResearcherAgent
 * A specialized agent for gathering and synthesizing information from the web.
 *
 * Architecture:
 * - Extends `SpecializedAgent` to participate in the Mesh (P2P agent network).
 * - Delegates actual research to `DeepResearchService`, which handles:
 *   1. Query generation via LLM
 *   2. Web search via `WebSearchTool` (DuckDuckGo)
 *   3. Recursive sub-topic exploration (depth/breadth controlled)
 *   4. LLM-powered synthesis of findings into a structured report
 *   5. Memory storage via `MemoryManager`
 *
 * Why it exists:
 * - Separates the "agent identity" (capabilities, mesh registration) from the
 *   "service logic" (research pipeline). The agent is the interface; the service
 *   is the engine.
 */
export class ResearcherAgent extends SpecializedAgent implements IA2AClient {
    /** The research engine that performs the actual web crawling, LLM synthesis, and memory storage */
    private deepResearchService: DeepResearchService;
    private onMessageCallback?: (message: A2AMessage) => void;

    constructor(deepResearchService: DeepResearchService) {
        // Register with capabilities ['research', 'search', 'summarization']
        // so the Director can route matching tasks to this agent
        super('Researcher', ['research', 'search', 'summarization']);
        this.deepResearchService = deepResearchService;

        // Register with A2A Broker
        a2aBroker.registerAgent('researcher', this);
        this.startA2AHeartbeat();
    }

    private startA2AHeartbeat() {
        setInterval(() => {
            a2aBroker.routeMessage({
                id: `heartbeat-researcher-${Date.now()}`,
                timestamp: Date.now(),
                sender: 'researcher',
                type: A2AMessageType.HEARTBEAT,
                payload: {}
            });
        }, 15000);
    }

    // A2A Implementation
    async sendMessage(message: A2AMessage): Promise<void> {
        if (message.type === A2AMessageType.TASK_NEGOTIATION) {
            // Respond with capability report (Phase 106: Negotiation)
            await a2aBroker.routeMessage({
                id: `a2a-bid-${Date.now()}`,
                timestamp: Date.now(),
                sender: 'researcher',
                recipient: message.sender,
                type: A2AMessageType.CAPABILITY_REPORT,
                payload: {
                    agentId: 'researcher',
                    capabilities: ['research', 'search', 'summarization'],
                    canHandle: true,
                    estimatedLatencyMs: 15000, // Research takes time
                    reasoning: 'Ready to perform recursive deep search.'
                },
                replyTo: message.id
            });
            return;
        }

        if (message.type === A2AMessageType.TASK_REQUEST) {
            const result = await this.handleTask({ task: message.payload.task });
            await a2aBroker.routeMessage({
                id: `a2a-${Date.now()}`,
                timestamp: Date.now(),
                sender: 'researcher',
                recipient: message.sender,
                type: A2AMessageType.TASK_RESPONSE,
                payload: result,
                replyTo: message.id
            });
        } else if (message.type === A2AMessageType.STATE_UPDATE && message.payload.action === 'REPORT_CAPABILITIES') {
            await a2aBroker.routeMessage({
                id: `a2a-cap-${Date.now()}`,
                timestamp: Date.now(),
                sender: 'researcher',
                type: A2AMessageType.STATE_UPDATE,
                payload: { 
                    capabilities: ['research', 'search', 'summarization', 'web-browsing'],
                    role: 'Web Researcher'
                }
            });
        }
    }

    onMessage(callback: (message: A2AMessage) => void): void {
        this.onMessageCallback = callback;
    }

    async delegateTask(task: A2ATask, recipient: string): Promise<A2ATask> {
        await a2aBroker.routeMessage({
            id: `a2a-task-${task.id}`,
            timestamp: Date.now(),
            sender: 'researcher',
            recipient: recipient,
            type: A2AMessageType.TASK_REQUEST,
            payload: { task: task.description, metadata: { taskId: task.id, priority: task.priority } }
        });
        task.status = 'in_progress';
        return task;
    }

    /**
     * Executes a research task from the Mesh/Director.
     *
     * @param offer - Task offer containing `task` (the research query string)
     * @returns Structured result with summary, sources, and the full recursive report
     */
    public async handleTask(offer: any): Promise<any> {
        console.log(`[ResearcherAgent] 🔍 Investigating query: ${offer.task}`);

        try {
            // Perform real recursive research
            // depth=2: research the topic, then research each sub-topic one more level
            // breadth=3: explore up to 3 related sub-topics per level
            const result = await this.deepResearchService.recursiveResearch(offer.task, 2, 3);

            console.log(`[ResearcherAgent] 📄 Report Ready: ${result.topic}`);

            return {
                status: 'completed',
                // Flatten sources into a simple { source, content } array for mesh consumers
                findings: result.sources.map(s => ({ source: s.title, content: s.url })),
                summary: result.summary,
                fullReport: result,
                // Phase 96: Execution Telemetry
                modelMetadata: result.modelMetadata
            };
        } catch (error: any) {
            console.error(`[ResearcherAgent] 💥 Research failed:`, error);
            throw error;
        }
    }
}

