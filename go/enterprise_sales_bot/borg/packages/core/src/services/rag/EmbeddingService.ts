/**
 * EmbeddingService.ts
 * 
 * Abstraction layer for generating vector embeddings from text.
 * Supports OpenAI (remote) and local models (via transformers.js or local execution).
 */

export interface EmbeddingProvider {
    embedText(text: string): Promise<number[]>;
    embedBatch(texts: string[]): Promise<number[][]>;
    getDimension(): number;
}

export class OpenAIEmbeddingProvider implements EmbeddingProvider {
    private apiKey: string;
    private model: string;

    constructor(apiKey: string, model: string = 'text-embedding-3-small') {
        this.apiKey = apiKey;
        this.model = model;
    }

    async embedText(text: string): Promise<number[]> {
        const res = await this.embedBatch([text]);
        return res[0];
    }

    async embedBatch(texts: string[]): Promise<number[][]> {
        const response = await fetch('https://api.openai.com/v1/embeddings', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${this.apiKey}`
            },
            body: JSON.stringify({
                input: texts,
                model: this.model
            })
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(`OpenAI Embedding Error: ${response.statusText} - ${error}`);
        }

        const data = await response.json();
        return data.data.map((item: any) => item.embedding);
    }

    getDimension(): number {
        return this.model === 'text-embedding-3-small' ? 1536 : 1536; // Example default
    }
}

export class LocalEmbeddingProvider implements EmbeddingProvider {
    private pipe: any = null;

    constructor() {
        // Initialization happens lazily
    }

    private async init() {
        if (!this.pipe) {
            try {
                // Dynamic import of transformers to avoid heavy load at startup
                const { pipeline, env } = await import('@xenova/transformers');
                // Use local only
                env.allowLocalModels = true;
                this.pipe = await pipeline('feature-extraction', 'Xenova/all-MiniLM-L6-v2', {
                    quantized: true
                });
            } catch (e: any) {
                console.error("Local embedding init failed. Ensure @xenova/transformers is installed.", e);
                throw e;
            }
        }
    }

    async embedText(text: string): Promise<number[]> {
        await this.init();
        const output = await this.pipe(text, { pooling: 'mean', normalize: true });
        return Array.from(output.data);
    }

    async embedBatch(texts: string[]): Promise<number[][]> {
        await this.init();
        const results = [];
        for (const text of texts) {
            const output = await this.pipe(text, { pooling: 'mean', normalize: true });
            results.push(Array.from(output.data) as number[]);
        }
        return results;
    }

    getDimension(): number {
        return 384; // MiniLM-L6-v2 dimension
    }
}

export class EmbeddingService {
    private provider: EmbeddingProvider;

    constructor(providerInit?: 'openai' | 'local', openaiKey?: string) {
        if (providerInit === 'openai' && openaiKey) {
            this.provider = new OpenAIEmbeddingProvider(openaiKey);
        } else {
            // Default to local to save costs and run offline
            this.provider = new LocalEmbeddingProvider();
        }
    }

    async embed(text: string): Promise<number[]> {
        return this.provider.embedText(text);
    }

    async embedBatch(texts: string[]): Promise<number[][]> {
        return this.provider.embedBatch(texts);
    }

    getDimension(): number {
        return this.provider.getDimension();
    }
}
