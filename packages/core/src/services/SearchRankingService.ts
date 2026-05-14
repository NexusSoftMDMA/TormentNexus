/**
 * @file SearchRankingService.ts
 * @module packages/core/src/services/SearchRankingService
 *
 * WHAT:
 * A service that interfaces with the Go sidecar's ranking engine.
 *
 * WHY:
 * Ensures consistent BM25/Cosine ranking results between the TypeScript
 * orchestrator and the high-performance Go sidecar.
 *
 * HOW:
 * Calls the Go sidecar's search and ranking API endpoints.
 */

import { Tool } from "@modelcontextprotocol/sdk/types.js";

export interface ScoredTool {
    name: string;
    description?: string;
    score: number;
    matchReason?: string;
}

export class SearchRankingService {
    private sidecarBaseUrl: string;

    constructor(sidecarBaseUrl: string = process.env.BORG_SIDECAR_URL || 'http://localhost:4300') {
        this.sidecarBaseUrl = sidecarBaseUrl;
    }

    /**
     * Ranks a set of tools using the Go sidecar's BM25 ranking engine.
     */
    async rankResults(query: string, tools: Tool[]): Promise<ScoredTool[]> {
        if (!query) {
            return tools.map(t => ({
                name: t.name,
                description: t.description,
                score: 0,
                matchReason: 'Default listing'
            }));
        }

        try {
            const url = new URL(`${this.sidecarBaseUrl}/api/mcp/tools/search`);
            url.searchParams.append('query', query);

            const response = await fetch(url.toString(), {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                return this.fallbackRank(query, tools);
            }

            const result = await response.json();
            if (result.success && Array.isArray(result.data)) {
                return result.data.map((item: any) => ({
                    name: item.name,
                    description: item.description,
                    score: item.score || 0,
                    matchReason: item.matchReason || 'Matched'
                }));
            }

            return this.fallbackRank(query, tools);
        } catch (error) {
            return this.fallbackRank(query, tools);
        }
    }

    private fallbackRank(query: string, tools: Tool[]): ScoredTool[] {
        const lowerQuery = query.toLowerCase();
        return tools
            .map(t => {
                let score = 0;
                if (t.name.toLowerCase().includes(lowerQuery)) score += 10;
                if (t.description?.toLowerCase().includes(lowerQuery)) score += 5;
                return {
                    name: t.name,
                    description: t.description,
                    score,
                    matchReason: score > 0 ? 'Simple keyword match' : 'No match'
                };
            })
            .filter(t => t.score > 0)
            .sort((a, b) => b.score - a.score);
    }
}

export const searchRankingService = new SearchRankingService();
