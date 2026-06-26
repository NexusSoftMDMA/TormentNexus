import { MeshService, SwarmMessage, SwarmMessageType } from './MeshService.js';

interface TaskOffer {
    task: string;
    requirements?: string[];
    tools?: string[];
    toolPolicy?: {
        allow?: string[];
        deny?: string[];
    };
    [key: string]: unknown;
}

function getErrorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
}

function parseTaskOffer(payload: unknown): TaskOffer | null {
    if (!payload || typeof payload !== 'object') {
        return null;
    }

    const record = payload as Record<string, unknown>;
    if (typeof record.task !== 'string') {
        return null;
    }

    if (
        record.requirements !== undefined &&
        (!Array.isArray(record.requirements) || !record.requirements.every(req => typeof req === 'string'))
    ) {
        return null;
    }

    if (record.tools !== undefined && (!Array.isArray(record.tools) || !record.tools.every(tool => typeof tool === 'string'))) {
        return null;
    }

    const toolPolicy = record.toolPolicy;
    if (toolPolicy !== undefined) {
        if (!toolPolicy || typeof toolPolicy !== 'object') {
            return null;
        }

        const toolPolicyRecord = toolPolicy as Record<string, unknown>;
        if (
            toolPolicyRecord.allow !== undefined &&
            (!Array.isArray(toolPolicyRecord.allow) || !toolPolicyRecord.allow.every(tool => typeof tool === 'string'))
        ) {
            return null;
        }

        if (
            toolPolicyRecord.deny !== undefined &&
            (!Array.isArray(toolPolicyRecord.deny) || !toolPolicyRecord.deny.every(tool => typeof tool === 'string'))
        ) {
            return null;
        }
    }

    return record as TaskOffer;
}

export abstract class SpecializedAgent {
    protected mesh: MeshService;
    public nodeId: string;
    public role: string;
    protected capabilities: string[] = [];

    constructor(role: string, capabilities: string[] = []) {
        this.mesh = new MeshService();
        this.nodeId = this.mesh.nodeId;
        this.role = role;
        this.capabilities = capabilities;

        this.initialize();
    }

    private initialize() {
        console.log(`[${this.role}] 🤖 Initializing Specialized Agent (Node: ${this.nodeId.slice(0, 8)}...)`);

        this.mesh.on('message', async (msg: SwarmMessage) => {
            // 1. Handle Capability Queries
            if (msg.type === SwarmMessageType.CAPABILITY_QUERY) {
                console.log(`[${this.role}] ❓ Received Capability Query from ${msg.sender.slice(0, 8)}...`);
                // Use sendResponse for RPC
                this.mesh.sendResponse(msg, SwarmMessageType.CAPABILITY_RESPONSE, {
                    role: this.role,
                    capabilities: this.capabilities,
                    load: 0
                });
                return;
            }

            // 2. Handle Task Offers (Bidding Phase)
            if (msg.type === SwarmMessageType.TASK_OFFER) {
                const offer = parseTaskOffer(msg.payload);
                if (!offer) {
                    console.warn(`[${this.role}] ⚠️ Ignoring malformed task offer from ${msg.sender.slice(0, 8)}...`);
                    return;
                }

                if (this.canHandle(offer)) {
                    console.log(`[${this.role}] ✋ Bidding for Task Offer: ${offer.task.slice(0, 30)}...`);
                    this.mesh.sendResponse(msg, SwarmMessageType.TASK_BID, {
                        task: offer.task,
                        originalTaskId: offer.originalTaskId || msg.id,
                        load: 0 // In the future, report actual CPU/Token load here 
                    });
                }
                return;
            }

            // 3. Handle Task Assignments (Execution Phase)
            if (msg.type === SwarmMessageType.TASK_ASSIGN) {
                if (msg.target !== this.nodeId) return; // Ignore if assigned to someone else

                const offer = parseTaskOffer(msg.payload);
                if (!offer) return;

                console.log(`[${this.role}] 🤝 Accepted Task Assignment: ${offer.task.slice(0, 30)}...`);
                this.mesh.sendResponse(msg, SwarmMessageType.TASK_ACCEPT, { task: offer.task });

                try {
                    const result = await this.handleTask(offer);
                    // Send Result as a Direct Message
                    this.mesh.sendDirect(msg.sender, SwarmMessageType.TASK_RESULT, {
                        originalTaskId: offer.originalTaskId || msg.id, // Support the original ID being passed down
                        result
                    });
                } catch (e: unknown) {
                    this.mesh.sendDirect(msg.sender, SwarmMessageType.TASK_RESULT, {
                        originalTaskId: offer.originalTaskId || msg.id,
                        error: getErrorMessage(e)
                    });
                }
            }
        });
    }

    public canHandle(offer: TaskOffer): boolean {
        if (offer.requirements) {
            return offer.requirements.every((req: string) => this.capabilities.includes(req) || this.role === req);
        }
        return true;
    }

    /**
     * Abstract method to be implemented by specific agents (Coder, Researcher)
     */
    protected abstract handleTask(offer: TaskOffer): Promise<unknown>;

    public async destroy() {
        this.mesh.destroy();
    }
}
