
import fs from 'fs/promises';
import { existsSync } from 'fs';
import path from 'path';

export interface GraphNode {
    id: string; // Unique ID (e.g., file path, symbol name)
    type: string; // 'file', 'function', 'class', 'import'
    metadata?: any;
}

export interface GraphEdge {
    source: string;
    target: string;
    relation: string; // 'imports', 'calls', 'inherits', 'defines'
    weight?: number;
}

/**
 * GraphMemory
 *
 * A simple, persistent Adjacency List graph for storing code relationships.
 * Stores logic connections that vector search misses (e.g., precise dependency trees).
 */
export class GraphMemory {
    private nodes: Map<string, GraphNode> = new Map();
    private edges: GraphEdge[] = [];
    private persistPath: string;
    private initialized: boolean = false;

    constructor(storageRoot: string = process.cwd()) {
        this.persistPath = path.join(storageRoot, '.tormentnexus', 'memory', 'knowledge_graph.json');

    }

    public async initialize() {
        if (this.initialized) return;

        try {
            if (existsSync(this.persistPath)) {
                const data = JSON.parse(await fs.readFile(this.persistPath, 'utf-8'));
                this.nodes = new Map(data.nodes.map((n: GraphNode) => [n.id, n]));
                this.edges = data.edges || [];
                console.log(`[GraphMemory] Loaded ${this.nodes.size} nodes and ${this.edges.length} edges.`);
            }
        } catch (e) {
            console.warn('[GraphMemory] Failed to load graph (starting fresh):', e);
        }

        this.initialized = true;
    }

    public async save() {
        const dir = path.dirname(this.persistPath);
        if (!existsSync(dir)) await fs.mkdir(dir, { recursive: true });

        const data = {
            nodes: Array.from(this.nodes.values()),
            edges: this.edges
        };

        await fs.writeFile(this.persistPath, JSON.stringify(data, null, 2));
    }

    public addNode(node: GraphNode) {
        if (!this.nodes.has(node.id)) {
            this.nodes.set(node.id, node);
        }
    }

    public addEdge(edge: GraphEdge) {
        // Prevent duplicates
        const exists = this.edges.some(e =>
            e.source === edge.source &&
            e.target === edge.target &&
            e.relation === edge.relation
        );

        if (!exists) {
            this.edges.push(edge);
        }
    }

    public getNeighbors(nodeId: string, relation?: string): GraphNode[] {
        const targets = this.edges
            .filter(e => e.source === nodeId && (!relation || e.relation === relation))
            .map(e => e.target);

        return targets
            .map(id => this.nodes.get(id))
            .filter((n): n is GraphNode => !!n);
    }

    public getIncoming(nodeId: string, relation?: string): GraphNode[] {
        const sources = this.edges
            .filter(e => e.target === nodeId && (!relation || e.relation === relation))
            .map(e => e.source);

        return sources
            .map(id => this.nodes.get(id))
            .filter((n): n is GraphNode => !!n);
    }

    public getSnapshot(): { nodes: GraphNode[], edges: GraphEdge[] } {
        return {
            nodes: Array.from(this.nodes.values()),
            edges: [...this.edges]
        };
    }
}
