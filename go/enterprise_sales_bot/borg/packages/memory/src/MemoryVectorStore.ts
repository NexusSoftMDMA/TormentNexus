import { IVectorStore } from './IVectorStore.js';

export class MemoryVectorStore implements IVectorStore {
    private db: any[] = [];
    private embedder: any;
    private initialized: boolean = false;

    constructor() { }

    async initialize() {
        if (this.initialized) return;
        const { pipeline } = await import('@xenova/transformers');
        console.log(`[MemoryVectorStore] Loading embedding model (Xenova/all-MiniLM-L6-v2) for in-memory store...`);
        this.embedder = await pipeline('feature-extraction', 'Xenova/all-MiniLM-L6-v2');
        this.initialized = true;
        console.log(`[MemoryVectorStore] Ready.`);
    }

    async createEmbeddings(text: string): Promise<number[]> {
        if (!this.initialized) await this.initialize();
        const output = await this.embedder(text, { pooling: 'mean', normalize: true });
        return Array.from(output.data);
    }

    async addMemory(content: string, metadata: any) {
        if (!this.initialized) await this.initialize();
        const vector = await this.createEmbeddings(content);
        this.db.push({
            id: metadata.id || `mem_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
            vector,
            text: content,
            ...metadata,
            timestamp: Date.now()
        });
    }

    async addDocuments(docs: any[]) {
        if (!this.initialized) await this.initialize();
        if (docs.length === 0) return;

        const processed = await Promise.all(docs.map(async d => {
            if (d.vector) return { ...d, timestamp: d.timestamp || Date.now() };
            const text = d.text || d.content;
            return {
                id: d.id || `doc_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
                ...d,
                vector: await this.createEmbeddings(text),
                timestamp: Date.now()
            };
        }));

        this.db.push(...processed);
    }

    async get(id: string) {
        return this.db.find(doc => doc.id === id) || null;
    }

    async delete(ids: string[]) {
        this.db = this.db.filter(doc => !ids.includes(doc.id));
    }

    async reset() {
        this.db = [];
    }

    async listDocuments(where?: string, limit: number = 100) {
        // Advanced SQL-like where clause parsing is out of scope for lightweight fallback,
        // but we can slice to limit
        return this.db.slice(0, limit);
    }

    async search(query: string, limit: number = 5, where?: string) {
        if (!this.initialized) await this.initialize();
        const queryVec = await this.createEmbeddings(query);

        // Simple cosine similarity since vectors are normalized (dot product)
        const scored = this.db.map(doc => {
            let score = 0;
            if (doc.vector && doc.vector.length === queryVec.length) {
                for (let i = 0; i < queryVec.length; i++) {
                    score += queryVec[i] * doc.vector[i];
                }
            }
            return { ...doc, _distance: 1 - score }; // LanceDB returns distance where smaller is better
        });

        scored.sort((a, b) => a._distance - b._distance);
        return scored.slice(0, limit);
    }
}
