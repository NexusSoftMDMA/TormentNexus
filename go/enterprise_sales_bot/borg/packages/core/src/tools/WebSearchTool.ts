/**
 * WebSearchTool
 *
 * Provides public web search capabilities via DuckDuckGo (duck-duck-scrape).
 * Used by DeepResearchService for external information gathering.
 *
 * Features:
 * - Exponential backoff retry on DDG rate-limiting ("anomaly detected")
 * - Configurable result limit (default: 5)
 * - Returns structured results with title, url, and snippet
 *
 * Rate-Limit Handling:
 * DDG aggressively rate-limits rapid requests. This tool implements retry
 * with exponential backoff (2s → 4s → 8s) to handle burst scenarios
 * like DeepResearchService's recursive topic exploration.
 */

import { search, SafeSearchType, SearchTimeType } from 'duck-duck-scrape';

/** Maximum number of retry attempts on rate-limit errors */
const MAX_RETRIES = 3;

/** Base delay in ms for exponential backoff (doubles each retry) */
const BASE_DELAY_MS = 2000;

/**
 * Helper to extract error messages from unknown error types.
 * Avoids unsafe `any` casts while still providing useful error strings.
 */
function getErrorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
}

/**
 * Checks if an error is a DuckDuckGo rate-limit error.
 * DDG returns "DDG detected an anomaly" when too many requests are made.
 */
function isRateLimitError(error: unknown): boolean {
    const msg = getErrorMessage(error);
    return msg.includes('anomaly') || msg.includes('too quickly');
}

/**
 * Sleep utility for backoff delays.
 */
function sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
}

export const WebSearchTool = {
    name: "search_web",
    description: "Search the public web using DuckDuckGo. Returns top results with titles, links, and snippets.",
    inputSchema: {
        type: "object",
        properties: {
            query: { type: "string", description: "The search query" },
            limit: { type: "number", description: "Max results (default: 5)" }
        },
        required: ["query"]
    },

    /**
     * Executes a web search with automatic retry on rate-limiting.
     *
     * @param args.query - The search query string
     * @param args.limit - Maximum number of results to return (default: 5)
     * @returns MCP-formatted content array with JSON search results
     */
    handler: async (args: { query: string, limit?: number }) => {
        const limit = args.limit || 5;

        // Retry loop with exponential backoff for DDG rate-limits
        for (let attempt = 0; attempt <= MAX_RETRIES; attempt++) {
            try {
                console.log(`[WebSearch] Searching for: ${args.query}${attempt > 0 ? ` (retry ${attempt}/${MAX_RETRIES})` : ''}`);

                const results = await search(args.query, {
                    safeSearch: SafeSearchType.MODERATE,
                    time: SearchTimeType.ALL,
                });

                // Extract and format top results
                const top = results.results.slice(0, limit).map(r => ({
                    title: r.title,
                    url: r.url,
                    snippet: r.description
                }));

                return {
                    content: [{
                        type: "text",
                        text: JSON.stringify(top, null, 2)
                    }]
                };

            } catch (error: unknown) {
                // If it's a rate-limit error and we have retries left, back off and retry
                if (isRateLimitError(error) && attempt < MAX_RETRIES) {
                    const delay = BASE_DELAY_MS * Math.pow(2, attempt); // 2s, 4s, 8s
                    console.warn(`[WebSearch] Rate-limited. Backing off ${delay}ms before retry ${attempt + 1}/${MAX_RETRIES}...`);
                    await sleep(delay);
                    continue;
                }

                // Non-retryable error or exhausted retries
                console.error("[WebSearch] Error:", error);
                return {
                    content: [{
                        type: "text",
                        text: `Error performing search: ${getErrorMessage(error)}`
                    }]
                };
            }
        }

        // Should never reach here, but TypeScript needs it
        return {
            content: [{
                type: "text",
                text: "Error: search exhausted all retries"
            }]
        };
    }
};
