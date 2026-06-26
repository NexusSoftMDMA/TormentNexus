/**
 * DebateProtocol.ts
 * 
 * Facilitates adversarial LLM execution. 
 * E.g., Claude writes the code, GPT tries to break it, Claude defends, Judge decides.
 * 
 * v2.7.35: Wired generateTurn and evaluateDebate to the Autopilot Council API
 * for real multi-model adversarial debate. Falls back to local stub if unavailable.
 */

import { EventEmitter } from 'events';
import { resolveSwarmOrchestratorBase } from './orchestrator-base.js';

export type DebateMode = 'standard' | 'adversarial';
export type DebateTopicType = 'general' | 'mission-plan';
export type DebateTurnStance = 'support' | 'critique' | 'judge';

export interface DebatePersonaProfile {
    key: 'proponent' | 'opponent' | 'judge';
    label: string;
    stance: DebateTurnStance;
    isAdversarial: boolean;
    objective: string;
}

export interface DebateTurn {
    round: number;
    persona: string;
    label: string;
    model: string;
    stance: DebateTurnStance;
    isAdversarial: boolean;
    argument: string;
    timestamp: number;
}

export interface DebateConfig {
    topic: string;
    proponentModel: string;
    opponentModel: string;
    judgeModel: string;
    maxRounds: number;
    mode?: DebateMode;
    topicType?: DebateTopicType;
    /** Base URL for the autopilot council server */
    councilUrl?: string;
}

export interface DebateResult {
    winner: string;
    summary: string;
    history: DebateTurn[];
    mode: DebateMode;
    topicType: DebateTopicType;
    personas: DebatePersonaProfile[];
}

export class DebateProtocol extends EventEmitter {
    private history: DebateTurn[] = [];
    private config: DebateConfig & { mode: DebateMode; topicType: DebateTopicType };
    private councilUrl: string | null;

    constructor(config: DebateConfig) {
        super();
        this.config = {
            ...config,
            mode: config.mode || 'standard',
            topicType: config.topicType || 'general'
        };
        this.councilUrl = resolveSwarmOrchestratorBase(config.councilUrl);
    }

    /**
     * Executes the debate loop: Proponent argues, Opponent rebuts, Judge evaluates.
     * Each turn is a real LLM call routed through the Autopilot Council's
     * per-supervisor endpoints.
     */
    public async conductDebate(): Promise<DebateResult> {
        this.emit('debate:started', this.config);

        let currentArgument = this.config.topic;

        for (let round = 1; round <= this.config.maxRounds; round++) {
            // 1. Proponent speaks — argues FOR the topic
            const proTurn = await this.generateTurn(round, 'Proponent', this.config.proponentModel, currentArgument);
            this.history.push(proTurn);
            this.emit('debate:turn', proTurn);

            // 2. Opponent rebuts — argues AGAINST the proponent's position
            const conTurn = await this.generateTurn(round, 'Opponent', this.config.opponentModel, proTurn.argument);
            this.history.push(conTurn);
            this.emit('debate:turn', conTurn);

            currentArgument = conTurn.argument; // Feed opponent's rebuttal into next round
        }

        // 3. Judge evaluates the entire transcript
        const judgement = await this.evaluateDebate();
        this.emit('debate:concluded', judgement);

        return judgement;
    }

    /**
     * Generates a single debate turn by calling the Autopilot Council's
     * per-supervisor chat endpoint. The supervisor name is mapped from
     * the model string (e.g., 'claude-3-5-sonnet' → 'Claude' supervisor).
     */
    private async generateTurn(round: number, persona: string, model: string, context: string): Promise<DebateTurn> {
        const profile = this.getPersonaProfile(persona);

        try {
            if (!this.councilUrl) {
                throw new Error('No TormentNexus Orchestrator base configured.');
            }
            const supervisorName = this.modelToSupervisor(model);
            const systemPrompt = this.buildSystemPrompt(profile);

            const res = await fetch(`${this.councilUrl}/api/supervisors/${supervisorName}/chat`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    message: `[Round ${round}] ${profile.label} — Respond to: "${context}"`,
                    systemPrompt,
                    context: this.history.map(t => `[R${t.round} ${t.persona}] ${t.argument}`).join('\n')
                }),
                signal: AbortSignal.timeout(30000)
            });

            if (res.ok) {
                const data = await res.json();
                const argument = data.response || data.message || data.text || '';
                if (argument) {
                    return {
                        round,
                        persona,
                        label: profile.label,
                        model,
                        stance: profile.stance,
                        isAdversarial: profile.isAdversarial,
                        argument,
                        timestamp: Date.now()
                    };
                }
            }
        } catch (err: any) {
            console.warn(`[Debate] Council supervisor call failed for ${persona} (${model}): ${err.message}`);
        }

        // Fallback: produce a basic turn noting the model was unreachable
        return {
            round,
            persona,
            label: profile.label,
            model,
            stance: profile.stance,
            isAdversarial: profile.isAdversarial,
            argument: `[Fallback] ${profile.label} using ${model} — council unavailable. Original context: "${context.substring(0, 100)}..."`,
            timestamp: Date.now()
        };
    }

    /**
     * Passes the full debate transcript to the Judge model for final evaluation.
     * Uses the council debate endpoint for a holistic multi-model judgement.
     */
    private async evaluateDebate(): Promise<DebateResult> {
        const transcript = this.history
            .map(t => `[Round ${t.round} — ${t.label} (${t.model})]\n${t.argument}`)
            .join('\n\n---\n\n');
        const opponentProfile = this.getPersonaProfile('Opponent');
        const proponentProfile = this.getPersonaProfile('Proponent');
        const personas = this.getPersonaProfiles();

        try {
            if (!this.councilUrl) {
                throw new Error('No TormentNexus Orchestrator base configured.');
            }
            const judgeSupervisor = this.modelToSupervisor(this.config.judgeModel);
            const res = await fetch(`${this.councilUrl}/api/supervisors/${judgeSupervisor}/chat`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    message: `You are the JUDGE. Evaluate this ${this.config.maxRounds}-round ${this.config.mode} debate on: "${this.config.topic}". Declare a winner (${proponentProfile.label} or ${opponentProfile.label}) and explain your reasoning in 2-3 sentences. Prioritize concrete operational rigor over rhetorical confidence.\n\nTranscript:\n${transcript}`,
                    systemPrompt: `You are an impartial judge evaluating a structured debate between two AI models. The debate mode is ${this.config.mode} and the topic type is ${this.config.topicType}. In adversarial mode, reward concrete red-team findings when they expose missing safeguards, brittle assumptions, or hidden dependencies. Respond with JSON: {"winner": "${proponentProfile.label}"|"${opponentProfile.label}", "summary": "your reasoning"}`
                }),
                signal: AbortSignal.timeout(30000)
            });

            if (res.ok) {
                const data = await res.json();
                const judgeText = data.response || data.message || data.text || '';

                // Try to parse structured JSON from the judge's response
                try {
                    const jsonMatch = judgeText.match(/\{[\s\S]*\}/);
                    if (jsonMatch) {
                        const parsed = JSON.parse(jsonMatch[0]);
                        return {
                            winner: parsed.winner || proponentProfile.label,
                            summary: parsed.summary || judgeText,
                            history: this.history,
                            mode: this.config.mode,
                            topicType: this.config.topicType,
                            personas
                        };
                    }
                } catch { /* JSON parse failed — use raw text */ }

                return {
                    winner: judgeText.toLowerCase().includes(opponentProfile.label.toLowerCase()) ? opponentProfile.label : proponentProfile.label,
                    summary: judgeText || `Judge (${this.config.judgeModel}) evaluated ${this.config.maxRounds} rounds.`,
                    history: this.history,
                    mode: this.config.mode,
                    topicType: this.config.topicType,
                    personas
                };
            }
        } catch (err: any) {
            console.warn(`[Debate] Judge evaluation failed: ${err.message}`);
        }

        // Fallback judgement
        return {
            winner: this.config.mode === 'adversarial' ? opponentProfile.label : proponentProfile.label,
            summary: `[Fallback] Judge (${this.config.judgeModel}) was unavailable. Defaulting to ${this.config.mode === 'adversarial' ? 'the adversarial critique' : 'the proponent'} to bias toward safer follow-up review.`,
            history: this.history,
            mode: this.config.mode,
            topicType: this.config.topicType,
            personas
        };
    }

    private getPersonaProfiles(): DebatePersonaProfile[] {
        return [
            this.getPersonaProfile('Proponent'),
            this.getPersonaProfile('Opponent'),
            this.getPersonaProfile('Judge')
        ];
    }

    private getPersonaProfile(persona: string): DebatePersonaProfile {
        if (persona === 'Proponent') {
            return {
                key: 'proponent',
                label: this.config.topicType === 'mission-plan' ? 'Plan Advocate' : 'Proponent',
                stance: 'support',
                isAdversarial: false,
                objective: this.config.topicType === 'mission-plan'
                    ? 'Defend the mission plan and justify why its sequencing, decomposition, and safeguards are sufficient.'
                    : 'Argue in favor of the position with concrete technical support.'
            };
        }

        if (persona === 'Judge') {
            return {
                key: 'judge',
                label: 'Judge',
                stance: 'judge',
                isAdversarial: false,
                objective: 'Evaluate both sides, prefer rigor over confidence, and explain the decision clearly.'
            };
        }

        return {
            key: 'opponent',
            label: this.config.mode === 'adversarial'
                ? (this.config.topicType === 'mission-plan' ? 'Red Team Critic' : 'Adversarial Opponent')
                : 'Opponent',
            stance: 'critique',
            isAdversarial: this.config.mode === 'adversarial',
            objective: this.config.mode === 'adversarial'
                ? 'Break the plan by finding edge cases, unsafe assumptions, hidden dependencies, and missing rollback or observability paths.'
                : 'Argue against the position and challenge the main assumptions.'
        };
    }

    private buildSystemPrompt(profile: DebatePersonaProfile): string {
        if (profile.key === 'proponent') {
            return this.config.topicType === 'mission-plan'
                ? 'You are the PLAN ADVOCATE in a structured debate about a proposed mission plan. Defend the plan, explain why the decomposition is sound, justify priorities, and note any existing safeguards. Address critiques directly. Keep your response under 300 words.'
                : 'You are the PROPONENT in this debate. Argue strongly IN FAVOR of the position. Be specific, cite technical reasoning, and address counterarguments. Keep your response under 300 words.';
        }

        if (this.config.mode === 'adversarial') {
            return this.config.topicType === 'mission-plan'
                ? 'You are the RED TEAM CRITIC in an adversarial review of a mission plan. Your job is to intentionally stress-test the plan, identify hidden dependencies, missing prerequisites, failure modes, unsafe assumptions, security issues, observability gaps, and rollback weaknesses. Prefer concrete breakpoints over generic disagreement. Keep your response under 300 words.'
                : 'You are the ADVERSARIAL OPPONENT in this debate. Attack the proposal by finding brittle assumptions, edge cases, security risks, missing evidence, and practical failure modes. Keep your response under 300 words.';
        }

        return 'You are the OPPONENT in this debate. Argue strongly AGAINST the position. Find weaknesses, propose alternatives, and challenge assumptions. Keep your response under 300 words.';
    }

    /**
     * Maps a model identifier string to the corresponding Autopilot Council supervisor name.
     * The council has supervisors: GPT-4o, Claude, DeepSeek, Gemini, Grok
     */
    private modelToSupervisor(model: string): string {
        const m = model.toLowerCase();
        if (m.includes('claude') || m.includes('anthropic')) return 'Claude';
        if (m.includes('gemini') || m.includes('google')) return 'Gemini';
        if (m.includes('deepseek')) return 'DeepSeek';
        if (m.includes('grok')) return 'Grok';
        return 'GPT-4o'; // Default
    }
}
