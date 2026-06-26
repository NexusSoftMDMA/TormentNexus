import { LLMService } from "@tormentnexus/ai";
import { RiskEvaluator, RiskResult } from "./RiskEvaluator.js";

export interface Participant {
    id: string;
    name: string;
    role: string;
    systemPrompt: string;
    model?: string;
    provider?: string;
}

export interface DebateTurn {
    participantId: string;
    participantName: string;
    role: string;
    content: string;
    timestamp: string;
    usage?: {
        prompt_tokens: number;
        completion_tokens: number;
        total_tokens: number;
    };
}

export interface DebateRound {
    roundNumber: number;
    turns: DebateTurn[];
}

export interface DebateResult {
    topic?: string;
    summary: string;
    rounds: DebateRound[];
    risk?: RiskResult;
    durationMs: number;
    totalUsage: {
        prompt_tokens: number;
        completion_tokens: number;
        total_tokens: number;
    };
}

export class DebateEngine {
    private riskEvaluator: RiskEvaluator;

    constructor(private llmService: LLMService) {
        this.riskEvaluator = new RiskEvaluator(llmService);
    }

    private async selectDebateModel(
        providerPreference: string | undefined,
        routingTaskType: 'coding' | 'general',
        taskType: 'worker' | 'supervisor',
    ): Promise<{ provider: string; modelId: string }> {
        return await this.llmService.modelSelector.selectModel({
            taskComplexity: taskType === 'worker' ? 'medium' : 'high',
            provider: providerPreference,
            taskType,
            routingTaskType,
        });
    }

    /**
     * Runs a multi-round debate between specialized agents.
     */
    async runDebate(topic: string, participants: Participant[], rounds: number = 1): Promise<DebateResult> {
        const startTime = Date.now();
        const debateRounds: DebateRound[] = [];
        const history: { role: string; content: string }[] = [];
        
        const totalUsage = {
            prompt_tokens: 0,
            completion_tokens: 0,
            total_tokens: 0
        };

        for (let i = 0; i < rounds; i++) {
            const turns: DebateTurn[] = [];
            for (const p of participants) {
                const systemPrompt = `You are ${p.name}, acting as a ${p.role}.
                TOPIC: ${topic}
                
                YOUR GUIDELINES:
                ${p.systemPrompt}

                Review the current discussion and provide your input. 
                Be critical but constructive. Focus on technical accuracy and system integrity.`;

                try {
                    const selection = p.provider && p.model
                        ? { provider: p.provider, modelId: p.model }
                        : await this.selectDebateModel(p.provider, 'coding', 'worker');
                    const response = await this.llmService.generateText(
                        selection.provider,
                        selection.modelId,
                        systemPrompt,
                        `Continue the debate on: ${topic}`,
                        {
                            history: [...history],
                            taskType: 'worker',
                            routingTaskType: 'coding' // Default to coding for technical debates
                        }
                    );

                    const turn: DebateTurn = {
                        participantId: p.id,
                        participantName: p.name,
                        role: p.role,
                        content: response.content,
                        timestamp: new Date().toISOString(),
                        usage: response.usage ? {
                            prompt_tokens: response.usage.inputTokens,
                            completion_tokens: response.usage.outputTokens,
                            total_tokens: response.usage.inputTokens + response.usage.outputTokens
                        } : undefined
                    };

                    if (turn.usage) {
                        totalUsage.prompt_tokens += turn.usage.prompt_tokens;
                        totalUsage.completion_tokens += turn.usage.completion_tokens;
                        totalUsage.total_tokens += turn.usage.total_tokens;
                    }

                    turns.push(turn);
                    history.push({ role: 'assistant', content: `[${p.name} (${p.role})]: ${response.content}` });
                } catch (error) {
                    console.error(`[Agents:Debate] Error in turn for ${p.name}:`, error);
                }
            }
            debateRounds.push({ roundNumber: i + 1, turns });
        }

        // Final Synthesis
        const synthesisPrompt = `You are the tormentnexus Collective Moderator. 
        Synthesize the following debate into a final consensus or recommendation.
        TOPIC: ${topic}
        
        DEBATE HISTORY:
        ${history.map(h => h.content).join('\n\n')}
        
        Provide a structured summary in Markdown including:
        1. Key Arguments
        2. Areas of Consensus
        3. Remaining Risks
        4. Final Recommendation`;

        const synthesisSelection = await this.selectDebateModel(undefined, 'general', 'supervisor');
        const synthesisResponse = await this.llmService.generateText(
            synthesisSelection.provider,
            synthesisSelection.modelId,
            'You are the tormentnexus Collective Moderator.',
            synthesisPrompt,
            {
                taskType: 'supervisor',
                routingTaskType: 'general',
            },
        );

        const result: DebateResult = {
            topic,
            summary: synthesisResponse.content,
            rounds: debateRounds,
            durationMs: Date.now() - startTime,
            totalUsage
        };

        // Attach risk analysis
        result.risk = await this.riskEvaluator.calculateRiskScore(topic, result.summary);

        return result;
    }

    /**
     * Runs a team conference (single-round sync).
     */
    async runConference(topic: string, participants: Participant[]): Promise<DebateResult> {
        return this.runDebate(topic, participants, 1);
    }
}
