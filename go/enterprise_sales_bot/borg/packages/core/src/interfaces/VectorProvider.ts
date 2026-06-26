export interface Document {
    id: string;
    content: string;
    metadata?: Record<string, any>;
    vector?: number[];
}

export interface SearchResult extends Document {
    score: number;
}

export interface Message {
    role: string;
    content: string | Array<{ type: string; text?: string;[key: string]: any }>;
    [key: string]: any;
}

export interface VectorProvider {
    /**
     * Initialize the provider (connect to DB, load models)
     */
    initialize(): Promise<void>;

    /**
     * Add documents to the store
     */
    add(docs: Document[]): Promise<void>;

    /**
     * Search for similar documents
     */
    search(query: string, limit?: number): Promise<SearchResult[]>;

    /**
     * Get by ID
     */
    get?(id: string): Promise<SearchResult | null>;

    /**
     * Delete documents by ID
     */
    delete(ids: string[]): Promise<void>;

    /**
     * Clear all data
     */
    reset(): Promise<void>;
}
