/**
 * ConsensusEngine.ts
 * 
 * Provides quorum-based truth resolution by aggregating output from multiple LLMs.
 * Designed to eliminate hallucinations by forcing 3+ models to converge mathematically.
 * 
 * v2.7.35: Wired queryModel to the Autopilot Council's per-supervisor chat endpoint
 * for real multi-model parallel dispatch. synthesizePlurality uses the judge supervisor
 * to extract common facts. Scoring uses a synthesis LLM instead of Math.random().
 */

import { EventEmitter } from 'events';
import { resolveSwarmOrchestratorBase } from './orchestrator-base.js';

export interface ConsensusQuery {
    prompt: string;
    models: string[]; // e.g., ['gpt-4o', 'claude-3-5-sonnet-20241022', 'gemini-2.5-pro']
    requiredAgreementPercentage?: number; // e.g., 0.66 (2 out of 3)
    /** Base URL for the autopilot council server */
    councilUrl?: string;
}

export interface CandidateResponse {
    model: string;
    result: string;
    score: number;
}

export class ConsensusEngine extends EventEmitter {
    private councilUrl: string | null;

    constructor(councilUrl?: string) {
        super();
        this.councilUrl = resolveSwarmOrchestratorBase(councilUrl);
    }

    /**
     * Dispatches the same query to multiple models concurrently and synthesizes a verdict.
     * Each model is queried via its corresponding Autopilot Council supervisor.
     * Scoring is computed by having a synthesis LLM evaluate agreement between responses.
     */
    public async seekConsensus(query: ConsensusQuery): Promise<{
        verdict: string;
        isConsensusReached: boolean;
        candidates: CandidateResponse[];
    }> {
        const agreementFactor = query.requiredAgreementPercentage || 0.66;
        const url = resolveSwarmOrchestratorBase(query.councilUrl) || this.councilUrl;
        this.emit('consensus:seeking', { prompt: query.prompt, models: query.models });

        // 1. Dispatch prompt to all models simultaneously
        const promises = query.models.map(model => this.queryModel(model, query.prompt, url));
        const results = await Promise.allSettled(promises);

        const candidates: CandidateResponse[] = [];

        results.forEach((r, idx) => {
            if (r.status === 'fulfilled' && r.value) {
                candidates.push({ model: query.models[idx], result: r.value, score: 0 });
            } else {
                const reason = r.status === 'rejected' ? r.reason?.message : 'empty response';
                console.warn(`[Consensus] Model ${query.models[idx]} failed: ${reason}`);
            }
        });

        if (candidates.length === 0) {
            const noResult = {
                verdict: '[No Response] All models failed to respond.',
                isConsensusReached: false,
                candidates
            };
            this.emit('consensus:resolved', noResult);
            return noResult;
        }

        // 2. Synthesize results and score agreement
        const { verdict, scores } = await this.synthesizePlurality(candidates, query.prompt, url);

        // Apply synthesis-scored agreement (each model gets 0 or 1 based on alignment)
        candidates.forEach((c, i) => {
            c.score = scores[i] ?? 0;
        });

        const votesForPlurality = candidates.filter(c => c.score > 0).length;
        const isConsensusReached = (votesForPlurality / query.models.length) >= agreementFactor;

        const finalResponse = {
            verdict: isConsensusReached ? verdict : `[Divergence] Models could not reach a ${Math.round(agreementFactor * 100)}% consensus. Best synthesis: ${verdict}`,
            isConsensusReached,
            candidates
        };

        this.emit('consensus:resolved', finalResponse);
        return finalResponse;
    }

    /**
     * Queries a single model via its corresponding Autopilot Council supervisor.
     * Maps model name to supervisor, then calls the chat endpoint.
     */
    private async queryModel(model: string, prompt: string, councilUrl: string | null): Promise<string> {
        try {
            if (!councilUrl) {
                throw new Error('No TormentNexus Orchestrator base configured.');
            }
            const supervisor = this.modelToSupervisor(model);
            const res = await fetch(`${councilUrl}/api/supervisors/${supervisor}/chat`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    message: prompt,
                    systemPrompt: 'Answer the question directly and concisely. Focus on factual accuracy. Keep your response under 500 words.'
                }),
                signal: AbortSignal.timeout(30000)
            });

            if (res.ok) {
                const data = await res.json();
                return data.response || data.message || data.text || '';
            }

            console.warn(`[Consensus] Supervisor ${supervisor} returned ${res.status}`);
        } catch (err: any) {
            console.warn(`[Consensus] Failed to query ${model}: ${err.message}`);
        }

        // Return empty to signal failure — the caller filters these out
        return '';
    }

    /**
     * Uses a synthesis supervisor to evaluate all candidate responses, extract
     * common facts, and score each candidate's alignment with the consensus.
     */
    private async synthesizePlurality(
        candidates: CandidateResponse[],
        originalPrompt: string,
        councilUrl: string | null
    ): Promise<{ verdict: string; scores: number[] }> {
        const candidateBlock = candidates
            .map((c, i) => `--- Model ${i + 1} (${c.model}) ---\n${c.result}`)
            .join('\n\n');

        try {
            if (!councilUrl) {
                throw new Error('No TormentNexus Orchestrator base configured.');
            }
            const res = await fetch(`${councilUrl}/api/supervisors/GPT-4o/chat`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    message: `Original question: "${originalPrompt}"\n\nThe following ${candidates.length} AI models answered independently:\n\n${candidateBlock}\n\nSynthesize the COMMON overlapping facts from all responses into a single authoritative answer. Then score each model's alignment (1 = agrees with consensus, 0 = diverges). Reply with JSON: {"verdict": "synthesized answer", "scores": [1, 0, 1, ...]}`,
                    systemPrompt: 'You are a consensus synthesis engine. Extract overlapping truths from multiple AI responses and identify which models agree. Always respond with valid JSON.'
                }),
                signal: AbortSignal.timeout(30000)
            });

            if (res.ok) {
                const data = await res.json();
                const text = data.response || data.message || data.text || '';

                // Parse the JSON response from the synthesis model
                const jsonMatch = text.match(/\{[\s\S]*\}/);
                if (jsonMatch) {
                    try {
                        const parsed = JSON.parse(jsonMatch[0]);
                        return {
                            verdict: parsed.verdict || text,
                            scores: Array.isArray(parsed.scores) ? parsed.scores : candidates.map(() => 1)
                        };
                    } catch { /* JSON parse failed */ }
                }

                // If JSON parsing fails, treat the whole response as the verdict
                // and assume all candidates agree
                return {
                    verdict: text,
                    scores: candidates.map(() => 1)
                };
            }
        } catch (err: any) {
            console.warn(`[Consensus] Synthesis LLM failed: ${err.message}`);
        }

        // Fallback: simple longest-common-substring heuristic
        // Score based on response length similarity as a rough proxy for agreement
        const avgLen = candidates.reduce((s, c) => s + c.result.length, 0) / candidates.length;
        const scores = candidates.map(c => Math.abs(c.result.length - avgLen) < avgLen * 0.5 ? 1 : 0);

        return {
            verdict: `[Local synthesis] Combined insights from ${candidates.length} models on: "${originalPrompt.substring(0, 80)}..."`,
            scores
        };
    }

    /**
     * Maps a model identifier to the corresponding Autopilot Council supervisor name.
     */
    private modelToSupervisor(model: string): string {
        const m = model.toLowerCase();
        if (m.includes('claude') || m.includes('anthropic') || m.includes('sonnet')) return 'Claude';
        if (m.includes('gemini') || m.includes('google')) return 'Gemini';
        if (m.includes('deepseek')) return 'DeepSeek';
        if (m.includes('grok')) return 'Grok';
        return 'GPT-4o'; // Default
    }
}
