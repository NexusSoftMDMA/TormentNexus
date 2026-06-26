import { connect } from '@lancedb/lancedb';
import { pipeline } from '@xenova/transformers';
import path from 'path';
import fs from 'fs';
import { IVectorStore } from './IVectorStore.js';

function sanitizeMetadataForArrow(metadata: Record<string, unknown>): Record<string, unknown> {
    const result: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(metadata)) {
        if (value === null || value === undefined) {
            result[key] = null;
        } else if (Array.isArray(value) || (typeof value === 'object')) {
            result[key] = JSON.stringify(value);
        } else {
            result[key] = value;
        }
    }
    return result;
}

// Global lock for table creation across all instances sharing the same DB path
const tableInitializationLocks = new Map<string, Promise<any>>();

export class LanceDBStore implements IVectorStore {
    private dbPath: string;
    private db: any;
    private embedder: any;
    private readonly HEAT_DECAY_HALFLIFE_MS = 1000 * 60 * 60 * 24;

    constructor(rootPath: string) {
        this.dbPath = path.resolve(rootPath, 'data', 'lancedb');
        if (!fs.existsSync(this.dbPath)) fs.mkdirSync(this.dbPath, { recursive: true });
    }

    async initialize() {
        this.db = await connect(this.dbPath);
        this.embedder = await pipeline('feature-extraction', 'Xenova/all-MiniLM-L6-v2');
    }

    private async ensureTable(initialData?: any[]) {
        const lockKey = this.dbPath;
        let promise = tableInitializationLocks.get(lockKey);

        if (!promise) {
            promise = (async () => {
                try {
                    return await this.db.openTable('memories');
                } catch (e) {
                    if (initialData && initialData.length > 0) {
                        try {
                            return await this.db.createTable('memories', initialData);
                        } catch (createErr: any) {
                            if (createErr.message?.includes('already exists')) {
                                return await this.db.openTable('memories');
                            }
                            throw createErr;
                        }
                    }
                    throw e;
                }
            })();
            tableInitializationLocks.set(lockKey, promise);
        }

        return promise;
    }

    async createEmbeddings(text: string): Promise<number[]> {
        const output = await this.embedder(text, { pooling: 'mean', normalize: true });
        return Array.from(output.data);
    }

    async addMemory(content: string, metadata: any) {
        const vector = await this.createEmbeddings(content);
        const { heat_score, last_accessed_at, timestamp, ...rest } = metadata;
        
        const data = [{ 
            vector, 
            text: content,
            heat_score: heat_score ?? 50, 
            last_accessed_at: last_accessed_at ?? Date.now(), 
            timestamp: timestamp ?? Date.now(),
            metadata: JSON.stringify(rest)
        }];
        const table = await this.ensureTable(data);
        await table.add(data);
    }

    async addDocuments(docs: any[]) {
        const processed = await Promise.all(docs.map(async d => {
            const { vector, text, content, heat_score, last_accessed_at, timestamp, metadata, ...rest } = d;
            const finalMetadata = { ...(metadata || {}), ...rest };
            
            return {
                vector: vector || await this.createEmbeddings(text || content),
                text: text || content,
                heat_score: heat_score ?? 50, 
                last_accessed_at: last_accessed_at ?? Date.now(), 
                timestamp: timestamp ?? Date.now(),
                metadata: JSON.stringify(finalMetadata)
            };
        }));
        const table = await this.ensureTable(processed);
        await table.add(processed);
    }

    async get(id: string) {
        try {
            const table = await this.ensureTable();
            const res = await table.search(await this.createEmbeddings('')).where(`id = '${id}'`).limit(1).toArray();
            return res.length > 0 ? res[0] : null;
        } catch (e) { return null; }
    }

    async delete(ids: string[]) {
        const table = await this.ensureTable();
        await table.delete(ids.map(id => `id = '${id}'`).join(' OR '));
    }

    async reset() { 
        tableInitializationLocks.delete(this.dbPath);
        await this.db.dropTable('memories'); 
    }

    async listDocuments(where?: string, limit: number = 100) {
        const table = await this.ensureTable();
        let q = table.search(await this.createEmbeddings('query')).limit(limit);
        if (where) q = q.where(where);
        return await q.toArray();
    }

    async search(query: string, limit: number = 5, where?: string) {
        const table = await this.ensureTable();
        let q = table.search(await this.createEmbeddings(query)).limit(limit);
        if (where) q = q.where(where);
        return await q.toArray();
    }

    async maintenance() {
        const table = await this.ensureTable();
        const all = await table.search(await this.createEmbeddings('')).limit(10000).toArray();
        const now = Date.now();
        const updates = all.map((item: any) => {
            const elapsed = now - (item.last_accessed_at || item.timestamp);
            const decay = Math.pow(0.5, elapsed / this.HEAT_DECAY_HALFLIFE_MS);
            return { ...item, heat_score: (item.heat_score || 50) * decay };
        });

        const maintenancePromise = (async () => {
            await this.db.dropTable('memories');
            return await this.db.createTable('memories', updates);
        })();

        tableInitializationLocks.set(this.dbPath, maintenancePromise);
        await maintenancePromise;
    }
}
