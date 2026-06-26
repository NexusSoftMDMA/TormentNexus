
import { DEFAULT_OPENROUTER_FREE_MODEL, LLMService } from '@tormentnexus/ai';
import type { MCPServer } from '../MCPServer.js';

export interface Mutation {
    id: string;
    originalPrompt: string;
    mutatedPrompt: string;
    reasoning: string;
    timestamp: number;
}

export interface Experiment {
    id: string;
    mutationId: string;
    task: string;
    status: 'PENDING' | 'RUNNING' | 'COMPLETED';
    resultA?: string;
    resultB?: string;
    winner?: 'A' | 'B' | 'TIE';
    judgeReasoning?: string;
}

interface LlmTextResponse {
    text?: string;
    content?: string;
}

/**
 * Reason: Darwin experiments consume outputs from multiple LLM calls that may vary in shape.
 * What: Collapses known response forms (`text`, `content`, raw string) into a stable string.
 * Why: Keeps mutation/judging logic deterministic while removing broad untyped cast usage.
 */
function extractLlmText(response: unknown): string {
    if (typeof response === 'string') {
        return response;
    }

    if (response && typeof response === 'object') {
        const maybeResponse = response as LlmTextResponse;
        if (typeof maybeResponse.text === 'string') {
            return maybeResponse.text;
        }
        if (typeof maybeResponse.content === 'string') {
            return maybeResponse.content;
        }
    }

    return String(response);
}

export class DarwinService {
    private llm: LLMService;
    private server: MCPServer;
    private mutations: Mutation[] = [];
    private experiments: Experiment[] = [];

    constructor(llm: LLMService, server: MCPServer) {
        this.llm = llm;
        this.server = server;
    }

    public async proposeMutation(originalPrompt: string, goal: string): Promise<Mutation> {
        const prompt = `
        You are The Evolutionary Engine.
        Your goal is to mutate the following system prompt to better achieve: "${goal}".
        Make it more concise, more strict, or more creative as needed.
        
        Original Prompt:
        "${originalPrompt}"
        
        Return JSON:
        {
            "mutatedPrompt": "The new prompt text",
            "reasoning": "Why this might be better"
        }
        `;

        const response = await this.llm.generateText("openrouter", DEFAULT_OPENROUTER_FREE_MODEL, "You are an expert prompt engineer. JSON only.", prompt, {});

        let result;
        try {
            result = JSON.parse(extractLlmText(response));
        } catch (e) {
            throw new Error("Failed to parse mutation suggestion.");
        }

        const mutation: Mutation = {
            id: Math.random().toString(36).substring(7),
            originalPrompt,
            mutatedPrompt: result.mutatedPrompt,
            reasoning: result.reasoning,
            timestamp: Date.now()
        };

        this.mutations.push(mutation);
        return mutation;
    }

    public async startExperiment(mutationId: string, task: string): Promise<Experiment> {
        const mutation = this.mutations.find(m => m.id === mutationId);
        if (!mutation) throw new Error("Mutation not found");

        const experiment: Experiment = {
            id: Math.random().toString(36).substring(7),
            mutationId,
            task,
            status: 'PENDING'
        };
        this.experiments.push(experiment);

        // Asynchronously run the experiment
        this.runExperimentLogic(experiment, mutation).catch(console.error);

        return experiment;
    }

    private async runExperimentLogic(experiment: Experiment, mutation: Mutation) {
        experiment.status = 'RUNNING';

        // 1. Run Agent A (Control)
        console.log(`[Darwin] Running Control (A) for Exp ${experiment.id}...`);
        // V1 Baseline execution
        const resultA = await this.runAgentExecution(mutation.originalPrompt, experiment.task);
        experiment.resultA = resultA;

        // 2. Run Agent B (Variant)
        console.log(`[Darwin] Running Variant (B) for Exp ${experiment.id}...`);
        const resultB = await this.runAgentExecution(mutation.mutatedPrompt, experiment.task);
        experiment.resultB = resultB;

        // 3. Judge
        console.log(`[Darwin] Judging Exp ${experiment.id}...`);
        const judgePrompt = `
        Compare two AI outputs for the task: "${experiment.task}".
        
        Output A (Control):
        ${resultA}
        
        Output B (Variant):
        ${resultB}
        
        Which is better? Return JSON:
        {
            "winner": "A" or "B" or "TIE",
            "reasoning": "Explanation"
        }
        `;

        const judgeRes = await this.llm.generateText("openrouter", DEFAULT_OPENROUTER_FREE_MODEL, "", judgePrompt, {});
        try {
            const verdict = JSON.parse(extractLlmText(judgeRes));
            experiment.winner = verdict.winner;
            experiment.judgeReasoning = verdict.reasoning;
        } catch (e) {
            experiment.winner = 'TIE';
            experiment.judgeReasoning = "Failed to parse judge verdict.";
        }

        experiment.status = 'COMPLETED';
        console.log(`[Darwin] Experiment ${experiment.id} Complete. Winner: ${experiment.winner}`);
    }

    private async runAgentExecution(sysPrompt: string, task: string): Promise<string> {
        // Run an agent executing a task with a specific system prompt
        // Using direct LLM invocation contextually
        const res = await this.llm.generateText("openrouter", DEFAULT_OPENROUTER_FREE_MODEL, sysPrompt, task, {});
        return extractLlmText(res);
    }

    public getStatus() {
        return {
            mutations: this.mutations,
            experiments: this.experiments
        };
    }
}
