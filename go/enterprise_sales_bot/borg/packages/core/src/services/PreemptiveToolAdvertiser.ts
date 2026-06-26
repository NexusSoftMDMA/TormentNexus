import { DEFAULT_OPENROUTER_FREE_MODEL, LLMService } from '@tormentnexus/ai';
import { SearchService, type SearchResult } from '@tormentnexus/search';

export class PreemptiveToolAdvertiser {
    private llmService: LLMService;
    private searchService: SearchService;

    constructor(llmService: LLMService, searchService: SearchService) {
        this.llmService = llmService;
        this.searchService = searchService;
    }

    /**
     * Analyzes the last few messages in a conversation and returns a string
     * advertising relevant tools preemptively if they match the topic.
     */
    public async getPreemptiveToolAdvertisements(recentMessages: { role: string; content: string }[]): Promise<string> {
        if (!recentMessages || recentMessages.length === 0) {
            return '';
        }

        const recentContext = recentMessages.slice(-3).map(m => `${m.role}: ${m.content}`).join('\n');
        
        // 1. Analyze topic
        const topicResponse = await this.llmService.generateText(
            'openrouter', 
            DEFAULT_OPENROUTER_FREE_MODEL,
            'You are a tool orchestration assistant. Determine the core actionable technical topic or task from the user\'s recent messages. Output ONLY the topic keywords.',
            `Recent messages:\n${recentContext}`
        );

        const topicKeywords = topicResponse.content?.trim();
        if (!topicKeywords || topicKeywords.length < 3) {
            return '';
        }

        // 2. Search the indexed workspace for the most relevant tool/context references
        const workspaceRoot = process.cwd();
        const searchResults = (await this.searchService.search(topicKeywords, workspaceRoot)).slice(0, 5);
        if (!searchResults || searchResults.length === 0) {
            return '';
        }

        // 3. Format advertisement
        const toolList = searchResults
            .map((result: SearchResult) => {
                const location = typeof result.line === 'number'
                    ? `${result.file}:${result.line}`
                    : result.file;
                return `- **${location}**: ${result.content}`;
            })
            .join('\n');
        
        return `[System Note: Based on your current topic, the following relevant workspace/tooling references were found and may help immediately:]\n${toolList}`;
    }
}
