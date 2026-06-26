/**
 * ConversationalToolInjector
 *
 * Maintains a sliding window of recent conversation turns and uses a small,
 * fast model (Gemma 12b via Ollama / cheapest routed model) to predict which
 * non-always-on tools from the full catalog the user is most likely to need
 * next. Predicted tools are automatically loaded into the session working set
 * without displacing always-on tools.
 *
 * Lifecycle:
 *   1. MCPServer calls `appendTurn(role, text)` after each conversation event.
 *   2. MCPServer calls `getPredictedToolNames(catalogSnapshot)` on each
 *      ListTools request — returns a deduplicated, throttled set of names.
 *   3. NativeSessionMetaTools.injectConversationalTools(names) loads them.
 */

import type { LLMService } from '@tormentnexus/ai';
import type { ModelSelector } from '@tormentnexus/ai';

export interface ConversationalToolInjectorOptions {
    /** Max recent turns to keep in the sliding window. Default: 8. */
    maxWindowTurns?: number;
    /** Minimum ms between LLM prediction calls. Default: 3000. */
    throttleMs?: number;
    /** Max tools to inject per prediction round. Default: 5. */
    maxInjectedTools?: number;
    /** Minimum token count in window before triggering prediction. Default: 30. */
    minWindowTokens?: number;
}

export interface ConversationTurn {
    role: 'user' | 'assistant' | 'tool';
    text: string;
    timestamp: number;
}

export interface CatalogEntry {
    name: string;
    description?: string | null;
    alwaysOn?: boolean;
    loaded?: boolean;
    serverTags?: string[];
    toolTags?: string[];
    semanticGroup?: string;
    keywords?: string[];
}

export interface PredictionResult {
    toolNames: string[];
    fromCache: boolean;
    reason: string;
}

const DEFAULT_MAX_WINDOW_TURNS = 8;
const DEFAULT_THROTTLE_MS = 3_000;
const DEFAULT_MAX_INJECTED_TOOLS = 5;
const DEFAULT_MIN_WINDOW_TOKENS = 30;

/**
 * Rough token estimate: split on whitespace.
 */
function roughTokenCount(text: string): number {
    return text.trim().split(/\s+/).length;
}

export class ConversationalToolInjector {
    private readonly maxWindowTurns: number;
    private readonly throttleMs: number;
    private readonly maxInjectedTools: number;
    private readonly minWindowTokens: number;
    private readonly window: ConversationTurn[] = [];

    private lastPredictionAt = 0;
    private lastPredictionResult: PredictionResult = {
        toolNames: [],
        fromCache: false,
        reason: 'initial',
    };

    // Tracks which names we already successfully injected in the last round
    // so we don't noisily re-log every ListTools call.
    private lastInjectedSet = new Set<string>();

    constructor(
        private readonly llmService: LLMService,
        private readonly modelSelector: ModelSelector,
        options: ConversationalToolInjectorOptions = {},
    ) {
        this.maxWindowTurns = options.maxWindowTurns ?? DEFAULT_MAX_WINDOW_TURNS;
        this.throttleMs = options.throttleMs ?? DEFAULT_THROTTLE_MS;
        this.maxInjectedTools = options.maxInjectedTools ?? DEFAULT_MAX_INJECTED_TOOLS;
        this.minWindowTokens = options.minWindowTokens ?? DEFAULT_MIN_WINDOW_TOKENS;
    }

    /**
     * Append a conversation turn to the sliding window.
     * Call this whenever the MCP server receives or emits a message.
     */
    public appendTurn(role: ConversationTurn['role'], text: string): void {
        if (!text || !text.trim()) {
            return;
        }

        this.window.push({ role, text: text.trim(), timestamp: Date.now() });

        // Keep window bounded
        while (this.window.length > this.maxWindowTurns) {
            this.window.shift();
        }
    }

    /**
     * Returns the combined text of the current window.
     */
    public getWindowSummary(): string {
        return this.window
            .map((turn) => `[${turn.role}] ${turn.text}`)
            .join('\n');
    }

    /**
     * Returns the rough token count of the current window.
     */
    public getWindowTokenCount(): number {
        return roughTokenCount(this.getWindowSummary());
    }

    /**
     * Core prediction method. Given the full tool catalog (minus always-on
     * tools, which are guaranteed), predicts which tools should be pre-loaded.
     *
     * Uses throttling and caching to avoid hammering the LLM on every request.
     */
    public async getPredictedToolNames(
        catalog: CatalogEntry[],
    ): Promise<PredictionResult> {
        const now = Date.now();
        const windowText = this.getWindowSummary();
        const windowTokens = this.getWindowTokenCount();

        // Not enough conversation yet to make a meaningful prediction.
        if (windowTokens < this.minWindowTokens) {
            return {
                toolNames: [],
                fromCache: false,
                reason: `window too small (${windowTokens} tokens < ${this.minWindowTokens} minimum)`,
            };
        }

        // Throttle: return cached result if we called recently.
        if (now - this.lastPredictionAt < this.throttleMs) {
            return {
                ...this.lastPredictionResult,
                fromCache: true,
            };
        }

        // Only offer non-always-on, non-already-loaded tools (exclude meta tools like search_tools etc.)
        const META_TOOL_NAMES = new Set([
            'search_tools', 'load_tool', 'get_tool_schema', 'get_tool_context',
            'unload_tool', 'list_loaded_tools', 'set_capacity',
            'get_eviction_history', 'clear_eviction_history', 'list_all_tools',
        ]);

        const candidates = catalog.filter(
            (t) => !t.alwaysOn && !t.loaded && !META_TOOL_NAMES.has(t.name),
        );

        if (candidates.length === 0) {
            const result: PredictionResult = {
                toolNames: [],
                fromCache: false,
                reason: 'no non-always-on candidate tools in catalog',
            };
            this.lastPredictionAt = now;
            this.lastPredictionResult = result;
            return result;
        }

        // Build a compact tool list for the prompt (limit to 200 for context budget)
        const compactCatalog = candidates.slice(0, 200)
            .map((t) => `${t.name}: ${(t.description ?? '').slice(0, 100)}`)
            .join('\n');

        const systemPrompt =
`You are a predictive tool routing assistant embedded in a developer AI system (TormentNexus).
Given the recent conversation window below, select up to ${this.maxInjectedTools} tools from the provided catalog that the user is MOST LIKELY to need next in this conversation.

Rules:
- Only select tools genuinely relevant to the user's apparent current task/direction.
- Return a JSON array of tool name strings ONLY, no explanation text.
- If no tools are clearly relevant, return an empty array [].
- Example valid responses: ["github__create_issue","filesystem__write_file"] or []

Available tools catalog (name: description):
${compactCatalog}`;

        const userPrompt = `Recent conversation window:\n${windowText}\n\nWhich tools should be pre-loaded?`;

        try {
            // Try sidecar first (Go kernel / Gemma 12b local model)
            const sidecarResult = await this.tryLocalModelPrediction(userPrompt, systemPrompt);
            if (sidecarResult) {
                const validNames = sidecarResult.filter(
                    (name) => catalog.some((t) => t.name === name),
                );
                const result: PredictionResult = {
                    toolNames: validNames.slice(0, this.maxInjectedTools),
                    fromCache: false,
                    reason: 'local-model (sidecar) prediction',
                };
                this.lastPredictionAt = now;
                this.lastPredictionResult = result;
                this.logIfChanged(result);
                return result;
            }
        } catch (err: any) {
            console.error('[ConversationalToolInjector] Sidecar prediction failed:', err?.message);
        }

        // Fallback: cheapest routed LLM
        try {
            const modelSelection = await this.modelSelector.selectModel({
                taskComplexity: 'low',
                routingTaskType: 'general',
                routingStrategy: 'cheapest',
            });

            const completion = await this.llmService.generateText(
                modelSelection.provider,
                modelSelection.modelId,
                systemPrompt,
                userPrompt,
                { routingStrategy: 'cheapest' },
            );

            const content = typeof completion === 'string' ? completion : (completion as any).content ?? '';
            const parsed = this.extractJsonArray(content);

            if (parsed !== null) {
                const validNames = parsed
                    .filter((name: string) => catalog.some((t) => t.name === name));
                const result: PredictionResult = {
                    toolNames: validNames.slice(0, this.maxInjectedTools),
                    fromCache: false,
                    reason: `cloud-model (${modelSelection.modelId}) prediction`,
                };
                this.lastPredictionAt = now;
                this.lastPredictionResult = result;
                this.logIfChanged(result);
                return result;
            }
        } catch (err: any) {
            console.error('[ConversationalToolInjector] Cloud-model prediction failed:', err?.message);
        }

        // Both failed — return empty to be safe.
        const fallback: PredictionResult = {
            toolNames: [],
            fromCache: false,
            reason: 'all predictors failed — no injection',
        };
        this.lastPredictionAt = now;
        this.lastPredictionResult = fallback;
        return fallback;
    }

    /**
     * Attempts to call the local sidecar (Go kernel) or Ollama Gemma endpoint
     * for fast, free predictions.
     */
    private async tryLocalModelPrediction(
        prompt: string,
        systemPrompt: string,
    ): Promise<string[] | null> {
        const SIDECAR_URL =
            process.env.TORMENTNEXUS_SIDECAR_URL ?? 'http://127.0.0.1:4300';
        const OLLAMA_URL =
            process.env.TORMENTNEXUS_OLLAMA_URL ?? 'http://127.0.0.1:11434';
        const LOCAL_MODEL =
            process.env.TORMENTNEXUS_LOCAL_PREDICT_MODEL ?? 'gemma3:12b';

        // 1. Try TormentNexus Go sidecar /api/mcp/tools/predict-conversational
        try {
            const res = await fetch(
                `${SIDECAR_URL}/api/mcp/tools/predict-conversational`,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ prompt, systemPrompt }),
                    signal: AbortSignal.timeout(5_000),
                },
            );

            if (res.ok) {
                const json = await res.json() as any;
                const tools = json?.data?.tools ?? json?.tools ?? json?.data ?? null;
                if (Array.isArray(tools)) {
                    return tools as string[];
                }
            }
        } catch {
            // sidecar not available — try Ollama
        }

        // 2. Try Ollama directly (Gemma 12b or configured model)
        try {
            const res = await fetch(`${OLLAMA_URL}/api/chat`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    model: LOCAL_MODEL,
                    stream: false,
                    messages: [
                        { role: 'system', content: systemPrompt },
                        { role: 'user', content: prompt },
                    ],
                    options: { temperature: 0.1, num_predict: 128 },
                }),
                signal: AbortSignal.timeout(8_000),
            });

            if (res.ok) {
                const json = await res.json() as any;
                const content: string = json?.message?.content ?? '';
                const parsed = this.extractJsonArray(content);
                if (parsed !== null) {
                    return parsed;
                }
            }
        } catch {
            // Ollama not available
        }

        return null;
    }

    /**
     * Safely extract the first JSON array from LLM text output.
     */
    private extractJsonArray(text: string): string[] | null {
        try {
            const start = text.indexOf('[');
            const end = text.lastIndexOf(']');
            if (start === -1 || end === -1) return null;
            const parsed = JSON.parse(text.slice(start, end + 1));
            if (Array.isArray(parsed) && parsed.every((item) => typeof item === 'string')) {
                return parsed;
            }
        } catch {
            // ignore parse errors
        }
        return null;
    }

    /**
     * Log only when the injected set changes to avoid log spam.
     */
    private logIfChanged(result: PredictionResult): void {
        const newSet = new Set(result.toolNames);
        const changed =
            newSet.size !== this.lastInjectedSet.size ||
            [...newSet].some((name) => !this.lastInjectedSet.has(name));

        if (changed && result.toolNames.length > 0) {
            console.error(
                `[ConversationalToolInjector] 🔮 Injecting tools via ${result.reason}: ${result.toolNames.join(', ')}`,
            );
        }

        this.lastInjectedSet = newSet;
    }

    /**
     * Returns a read-only snapshot of the current sliding window.
     */
    public getWindow(): ReadonlyArray<ConversationTurn> {
        return this.window;
    }

    /**
     * Clear the window (e.g., on session reset).
     */
    public clearWindow(): void {
        this.window.length = 0;
        this.lastPredictionAt = 0;
        this.lastInjectedSet.clear();
        this.lastPredictionResult = {
            toolNames: [],
            fromCache: false,
            reason: 'window cleared',
        };
    }
}
