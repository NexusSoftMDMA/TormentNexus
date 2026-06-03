
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { GraphMemory } from '@tormentnexus/memory';
import fs from 'fs/promises';
import path from 'path';

describe('Phase 27: Logic Graph (Knowledge Graph)', () => {
    const testRoot = path.join(process.cwd(), '.tormentnexus_test_graph');

    beforeEach(async () => {
        await fs.mkdir(testRoot, { recursive: true });
    });

    afterEach(async () => {
        await fs.rm(testRoot, { recursive: true, force: true });
    });

    it('should add nodes and edges and retrieve neighbors', async () => {
        const graph = new GraphMemory(testRoot);
        await graph.initialize();

        graph.addNode({ id: 'A', type: 'file' });
        graph.addNode({ id: 'B', type: 'file' });
        graph.addNode({ id: 'C', type: 'file' });

        graph.addEdge({ source: 'A', target: 'B', relation: 'imports' });
        graph.addEdge({ source: 'B', target: 'C', relation: 'imports' });

        const neighborsOfA = graph.getNeighbors('A');
        expect(neighborsOfA).toHaveLength(1);
        expect(neighborsOfA[0].id).toBe('B');

        const incomingToC = graph.getIncoming('C');
        expect(incomingToC).toHaveLength(1);
        expect(incomingToC[0].id).toBe('B');
    });

    it('should persist graph to disk', async () => {
        const graph = new GraphMemory(testRoot);
        await graph.initialize();
        graph.addNode({ id: 'PersistNode', type: 'test' });
        await graph.save();

        // Reload
        const graph2 = new GraphMemory(testRoot);
        await graph2.initialize();
        const neighbors = graph2.getNeighbors('PersistNode'); // Should be empty but node exists
        // actually getNeighbors returns targets. We need to check if node exists.
        // My simple implementation doesn't expose `hasNode` but `addNode` is idempotent.

        // Let's add an edge to verify persistence better
        graph.addNode({ id: 'Target', type: 'test' });
        graph.addEdge({ source: 'PersistNode', target: 'Target', relation: 'links' });
        await graph.save();

        const graph3 = new GraphMemory(testRoot);
        await graph3.initialize();
        const targets = graph3.getNeighbors('PersistNode');
        expect(targets).toHaveLength(1);
        expect(targets[0].id).toBe('Target');
    });
});
