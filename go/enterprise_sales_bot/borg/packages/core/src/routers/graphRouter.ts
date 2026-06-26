import { z } from 'zod';
import { t, publicProcedure } from '../lib/trpc-core.js';
import { RepoGraphService } from '../services/RepoGraphService.js';
import { getAutoTestService, getMemoryManager } from '../lib/trpc-core.js';

export interface SymbolGraphNode {
    id: string;
    name: string;
    val: number;
    group: 'symbol';
    kind?: string;
    file?: string;
}

export interface SymbolGraphLink {
    source: string;
    target: string;
    type: 'defines';
}

interface SymbolRecord {
    id?: string;
    metadata?: {
        name?: string;
        kind?: string;
        file_path?: string;
    };
}

// Lazy-initialized graph service (initialized on first call)
let graphService: RepoGraphService | null = null;

function getGraphService(): RepoGraphService {
    if (!graphService) {
        const repoGraph = getAutoTestService().repoGraph as RepoGraphService | undefined;
        if (repoGraph) {
            graphService = repoGraph;
        } else {
            // Fallback: create standalone instance
            graphService = new RepoGraphService(process.cwd());
        }
    }
    return graphService!;
}

export const graphRouter = t.router({
    get: publicProcedure.query(async () => {
        const service = getGraphService();
        if (!service) {
            return { nodes: [], links: [], dependencies: {} as Record<string, string[]> };
        }
        // Build graph if not initialized
        const isInitialized = Reflect.get(service, 'isInitialized') as boolean | undefined;
        if (!isInitialized) {
            await service.buildGraph();
        }
        return service.toJSON();
    }),

    rebuild: publicProcedure.mutation(async () => {
        const service = getGraphService();
        await service.buildGraph();
        return { success: true, ...service.toJSON() };
    }),

    getConsumers: publicProcedure
        .input(z.object({ filePath: z.string() }))
        .query(({ input }) => {
            const service = getGraphService();
            return service.getConsumers(input.filePath);
        }),

    getDependencies: publicProcedure
        .input(z.object({ filePath: z.string() }))
        .query(({ input }) => {
            const service = getGraphService();
            return service.getDependencies(input.filePath);
        }),

    getSymbolsGraph: publicProcedure.query(async () => {
        const symbols = await getMemoryManager().getAllSymbols?.() ?? [];

        const nodes: SymbolGraphNode[] = [];
        const links: SymbolGraphLink[] = [];

        symbols.forEach((sym: unknown) => {
            const symbol = sym as SymbolRecord;
            if (!symbol.id) {
                return;
            }
            const metadata = symbol.metadata ?? {};

            // Node for the symbol
            nodes.push({
                id: symbol.id,
                name: metadata.name ?? 'unknown-symbol',
                val: 5, // Size
                group: 'symbol',
                kind: metadata.kind, // function, class, etc.
                file: metadata.file_path
            });

            // Link to the file
            if (metadata.file_path) {
                links.push({
                    source: metadata.file_path,
                    target: symbol.id,
                    type: 'defines'
                });
            }
        });

        return { nodes, links };
    }),
});

