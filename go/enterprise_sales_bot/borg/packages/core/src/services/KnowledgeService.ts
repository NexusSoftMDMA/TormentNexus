
import { MemoryManager } from './MemoryManager.js';
import { GraphNode, GraphEdge } from '@tormentnexus/memory';

export interface ContextBundle {
    root: { id: string, type: string };
    nodes: GraphNode[];
    edges: GraphEdge[];
    depth: number;
}

export interface KnowledgeGraphContentBlock {
    type: 'text';
    text: string;
}

export interface KnowledgeGraphResponse {
    content: KnowledgeGraphContentBlock[];
}

interface SnapshotCapableGraph {
    getSnapshot(): unknown;
}

export class KnowledgeService {
    private memory: MemoryManager;

    /**
     * Reason: GraphMemory's public type does not currently guarantee `getSnapshot`, but some runtimes expose it.
     * What: Runtime guard that checks for snapshot capability without unsafe structural casts.
     * Why: Allows optional snapshot fast-path while preserving type safety and compatibility.
     */
    private isSnapshotCapableGraph(graph: unknown): graph is SnapshotCapableGraph {
        if (!graph || typeof graph !== 'object') {
            return false;
        }
        const getSnapshot = Reflect.get(graph as object, 'getSnapshot');
        return typeof getSnapshot === 'function';
    }

    constructor(memory: MemoryManager) {
        this.memory = memory;
    }

    /**
     * Retrieves the Knowledge Graph (generic)
     */
    public async getGraph(query?: string, depth: number = 1): Promise<KnowledgeGraphResponse> {
        // If no query, return full graph
        if (!query && this.memory.graph && this.isSnapshotCapableGraph(this.memory.graph)) {
            const snapshot = this.memory.graph.getSnapshot();
            return {
                content: [{
                    type: "text",
                    text: JSON.stringify(snapshot, null, 2)
                }]
            };
        }

        if (query) {
            const context = await this.getDeepContext(query, depth);
            return {
                content: [{
                    type: "text",
                    text: JSON.stringify(context, null, 2)
                }]
            };
        }

        return { content: [{ type: "text", text: "Please provide a query to search the graph." }] };
    }

    /**
     * Recursive Context Retrieval
     */
    public async getDeepContext(rootId: string, maxDepth: number = 2): Promise<ContextBundle> {
        if (!this.memory.graph) {
            console.warn("[KnowledgeService] Graph memory not available.");
            return { root: { id: rootId, type: 'unknown' }, nodes: [], edges: [], depth: 0 };
        }

        const visited = new Set<string>();
        const queue: { id: string, depth: number }[] = [{ id: rootId, depth: 0 }];
        const resultNodes: GraphNode[] = [];
        const resultEdges: GraphEdge[] = [];

        while (queue.length > 0) {
            const current = queue.shift()!;
            if (visited.has(current.id)) continue;
            visited.add(current.id);

            // Add current node to results? 
            // We need to fetch it from graph if possible, or just stub it
            // GraphMemory doesn't expose getNode(id) directly in the interface yet (only getNeighbors)
            // But we can infer it or we might need to update GraphMemory to support `getNode(id)`.
            // For now, let's just use neighbors to populate.

            if (current.depth >= maxDepth) continue;

            const neighbors = this.memory.graph.getNeighbors(current.id);

            for (const neighbor of neighbors) {
                // Track Edge (We don't have edge metadata directly from getNeighbors in the simple API, 
                // we might need to update GraphMemory again to return Edges, not just Nodes)
                // Assuming simple relation standard connections.

                resultEdges.push({
                    source: current.id,
                    target: neighbor.id,
                    relation: 'related' // Simplification
                });

                if (!visited.has(neighbor.id)) {
                    resultNodes.push(neighbor);
                    queue.push({ id: neighbor.id, depth: current.depth + 1 });
                }
            }
        }

        return {
            root: { id: rootId, type: 'inferred' },
            nodes: resultNodes,
            edges: resultEdges,
            depth: maxDepth
        };
    }
}
