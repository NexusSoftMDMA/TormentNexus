
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { KnowledgeService } from '../services/KnowledgeService.js';
import { MemoryManager } from '../services/MemoryManager.js';
import { GraphMemory } from '@tormentnexus/memory';

// Mock MemoryManager
class MockMemoryManager {
    graph: GraphMemory;
    constructor(graph: GraphMemory) {
        this.graph = graph;
    }
}

describe('Phase 28: Smart Context Retrieval', () => {
    let graph: GraphMemory;
    let service: KnowledgeService;
    const testPath = '.tormentnexus_test_smart_context';

    beforeEach(async () => {
        // Cleanup handled by GraphMemory logic usually, but here we use in-memory or temp path
        graph = new GraphMemory(testPath);
        await graph.initialize();

        // Seed Graph
        graph.addNode({ id: 'Root', type: 'file' });
        graph.addNode({ id: 'ChildA', type: 'file' });
        graph.addNode({ id: 'ChildB', type: 'file' });
        graph.addNode({ id: 'GrandChild', type: 'file' });

        graph.addEdge({ source: 'Root', target: 'ChildA', relation: 'imports' });
        graph.addEdge({ source: 'Root', target: 'ChildB', relation: 'imports' });
        graph.addEdge({ source: 'ChildA', target: 'GrandChild', relation: 'calls' });

        const memory = new MockMemoryManager(graph) as unknown as MemoryManager;
        service = new KnowledgeService(memory);
    });

    it('should retrieve direct neighbors (depth 1)', async () => {
        const result = await service.getDeepContext('Root', 1);
        expect(result.nodes.map(n => n.id)).toContain('ChildA');
        expect(result.nodes.map(n => n.id)).toContain('ChildB');
        expect(result.nodes.map(n => n.id)).not.toContain('GrandChild'); // Depth limit
    });

    it('should retrieve deep neighbors (depth 2)', async () => {
        const result = await service.getDeepContext('Root', 2);
        expect(result.nodes.map(n => n.id)).toContain('GrandChild');
    });
});
