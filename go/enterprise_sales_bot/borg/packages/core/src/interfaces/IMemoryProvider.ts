export interface Memory {
    uuid: string;
    content: string;
    embedding?: number[];
    metadata: Record<string, any>;
    userId: string;
    agentId?: string;
    createdAt: Date;
}

export interface IMemoryProvider {
    init(): Promise<void>;
    saveMemory(content: string, metadata: Record<string, any>, userId: string, agentId?: string): Promise<Memory>;
    searchMemories(query: string, userId: string, limit?: number, threshold?: number): Promise<Memory[]>;
    listMemories(userId: string, limit?: number, offset?: number): Promise<Memory[]>;
    deleteMemory(uuid: string, userId: string): Promise<void>;
}
