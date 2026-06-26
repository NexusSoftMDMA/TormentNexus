
import { v4 as uuidv4 } from 'uuid';
import { DEFAULT_OPENROUTER_FREE_MODEL, LLMService } from '@tormentnexus/ai';
import fs from 'fs';
import path from 'path';

export interface Suggestion {
    id: string;
    title: string;
    description: string;
    type: 'ACTION' | 'INFO';
    source: string; // e.g., "Director", "Linter"
    payload?: unknown; // Tool call data or other metadata
    timestamp: number;
    status: 'PENDING' | 'APPROVED' | 'REJECTED';
}

import { Director } from '@tormentnexus/agents';

interface LlmTextResponse {
    text?: string;
}

/**
 * Reason: LLM provider responses can be either plain strings or object payloads.
 * What: Normalizes supported response variants into a single text string.
 * Why: Keeps suggestion parsing stable while removing broad response casts.
 */
function extractLlmText(response: unknown): string {
    if (typeof response === 'string') {
        return response;
    }

    if (response && typeof response === 'object') {
        const typedResponse = response as LlmTextResponse;
        if (typeof typedResponse.text === 'string') {
            return typedResponse.text;
        }
    }

    return '';
}

export class SuggestionService {
    private suggestions: Suggestion[] = [];
    private llmService: LLMService;
    private processingPaths = new Set<string>(); // Prevent duplicate processing
    private persistencePath: string;
    private director?: Director;

    constructor(llmService?: LLMService, director?: Director) {
        this.llmService = llmService || new LLMService();
        this.director = director;
        this.persistencePath = path.join(process.cwd(), 'packages/core/data/suggestions.json');
        this.loadSuggestions();
    }

    private loadSuggestions() {
        try {
            if (fs.existsSync(this.persistencePath)) {
                const data = fs.readFileSync(this.persistencePath, 'utf-8');
                this.suggestions = JSON.parse(data);
                console.log(`[SuggestionService] Loaded ${this.suggestions.length} suggestions.`);
            }
        } catch (e) {
            console.error("[SuggestionService] Failed to load suggestions:", e);
        }
    }

    private saveSuggestions() {
        try {
            const dir = path.dirname(this.persistencePath);
            if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });

            // Prune old resolved items to keep file small (keep last 50)
            const active = this.suggestions.filter(s => s.status === 'PENDING');
            const history = this.suggestions.filter(s => s.status !== 'PENDING').slice(-50);

            fs.writeFileSync(this.persistencePath, JSON.stringify([...active, ...history], null, 2));
        } catch (e) {
            console.error("[SuggestionService] Failed to save suggestions:", e);
        }
    }

    /**
     * Creates a new suggestion and adds it to the queue.
     */
    addSuggestion(title: string, description: string, source: string, payload?: unknown): Suggestion {
        const suggestion: Suggestion = {
            id: uuidv4(),
            title,
            description,
            type: 'ACTION',
            source,
            payload,
            timestamp: Date.now(),
            status: 'PENDING'
        };
        this.suggestions.push(suggestion);
        this.saveSuggestions();
        console.log(`[SuggestionService] New Suggestion: ${title} (${source})`);

        // Conversational Nudge
        if (this.director) {
            // Fire and forget
            this.director.broadcast('suggestion', `💡 **Suggestion**: I found a potential improvement: **"${title}"**.\n${description}\n*(Check Dashboard to Approve)*`);
        }

        return suggestion;
    }

    getPendingSuggestions(): Suggestion[] {
        return this.suggestions.filter(s => s.status === 'PENDING').sort((a, b) => b.timestamp - a.timestamp);
    }

    clearAll() {
        this.suggestions = [];
        this.saveSuggestions();
    }

    resolveSuggestion(id: string, status: 'APPROVED' | 'REJECTED'): Suggestion | undefined {
        const index = this.suggestions.findIndex(s => s.id === id);
        if (index !== -1) {
            this.suggestions[index].status = status;
            this.saveSuggestions(); // Save state change
            console.log(`[SuggestionService] Suggestion ${id} ${status}`);
            return this.suggestions[index];
        }
        return undefined;
    }

    /**
     * Analyzes conversation history using LLM to generate intelligent suggestions.
     */
    async processConversation(history: Array<{ role: string; content: string }>) {
        if (history.length < 2) return;

        try {
            const contextText = history.slice(-5).map(h => `${h.role}: ${h.content}`).join('\n');

            const prompt = `
            You are a Neural OS Watcher. You are monitoring a conversation between a human and an agent.
            
            Conversation History (Last 5 turns):
            ${contextText}
            
            Analyze the semantic intent and thematic progression. If you predict the user will need a specific tool or skill next (e.g. searching the web, running code, saving to memory, or a specific domain skill), suggest it.
            
            Return JSON ONLY:
            {
                "found": boolean,
                "title": "Short Title",
                "description": "Why this tool/skill is relevant now",
                "tool": "tool_name_or_skill_id",
                "args": { ... }
            }
            
            If nothing specific is predicted, return { "found": false }.
            `;

            const response = await this.llmService.generateText(
                'openrouter',
                DEFAULT_OPENROUTER_FREE_MODEL,
                'You are a predictive intelligence agent.',
                prompt,
                { routingStrategy: 'cheapest' }
            );
            const textContent = extractLlmText(response);
            const jsonStart = textContent.indexOf('{');
            const jsonEnd = textContent.lastIndexOf('}');
            if (jsonStart !== -1 && jsonEnd !== -1) {
                const result = JSON.parse(textContent.slice(jsonStart, jsonEnd + 1));

                if (result.found) {
                    this.addSuggestion(
                        result.title,
                        result.description,
                        "Predictive Intelligence",
                        { tool: result.tool, args: result.args }
                    );
                }
            }
        } catch (e) {
            console.error("[SuggestionService] Conversation analysis failed:", e);
        }
    }

    /**
     * Analyzes context using LLM to generate intelligent suggestions.
     */
    async processContext(context: { type: string; path?: string; content?: string }) {
        if (!context.path || this.processingPaths.has(context.path)) return;

        // Cooldown/Debounce check
        this.processingPaths.add(context.path);
        setTimeout(() => this.processingPaths.delete(context.path!), 30000); // 30s cooldown per file

        try {
            // Keep content short to save tokens
            const snippet = context.content ? context.content.slice(0, 2000) : `(Path: ${context.path})`;

            const prompt = `
            You are an AI Pair Programmer. The user is editing: ${context.path}
            
            Code Snippet (First 2kb):
            ${snippet}
            
            Analyze this context. If you see a clear, actionable improvement (e.g., adding tests, fixing a security issue, refactoring complex code, or running a specific command), propose it.
            
            Return JSON ONLY:
            {
                "found": boolean,
                "title": "Short Title",
                "description": "Why checking this is important",
                "tool": "tool_name",
                "args": { ... }
            }
            
            If nothing important triggers, return { "found": false }.
            Do not suggest generic things like 'Add comments'. Only actionable, high-value tasks.
            `;

            // Simplified LLM call using defaults
            const response = await this.llmService.generateText(
                'openrouter',
                DEFAULT_OPENROUTER_FREE_MODEL,
                'You are an expert pair programmer analyzing code context.',
                prompt
            );
            const textContent = extractLlmText(response);
            const jsonStart = textContent.indexOf('{');
            const jsonEnd = textContent.lastIndexOf('}');
            if (jsonStart !== -1 && jsonEnd !== -1) {
                const result = JSON.parse(textContent.slice(jsonStart, jsonEnd + 1));

                if (result.found) {
                    this.addSuggestion(
                        result.title,
                        result.description,
                        "AI Copilot",
                        { tool: result.tool, args: result.args }
                    );
                }
            }
        } catch (e) {
            console.error("[SuggestionService] Analysis failed:", e);
        }
    }
}
