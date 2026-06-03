import { LLMService } from "@tormentnexus/ai";

export interface RiskResult {
    score: number;
    level: 'low' | 'medium' | 'high' | 'critical';
    rationale: string;
}

export class RiskEvaluator {
    constructor(private llmService: LLMService) {}

    private async selectRiskModel(routingTaskType: 'general' | 'planning'): Promise<{ provider: string; modelId: string }> {
        return await this.llmService.modelSelector.selectModel({
            taskComplexity: 'medium',
            taskType: 'supervisor',
            routingTaskType,
        });
    }

    /**
     * Calculates a risk score (0-100) for a proposed set of changes or debate outcome.
     */
    async calculateRiskScore(topic: string, summary: string): Promise<RiskResult> {
        const prompt = `
            Analyze the following technical task/debate result and provide a risk score between 0 and 100.
            
            SCORING CRITERIA:
            - 0-20 (LOW): Documentation, minor refactors, safe additions, high consensus.
            - 21-50 (MEDIUM): New features, UI changes, non-critical logic updates.
            - 51-80 (HIGH): Core logic changes, complex refactors, security-adjacent work.
            - 81-100 (CRITICAL): Database schema changes, auth/security logic, destructive operations, no consensus.

            TOPIC: ${topic}
            SUMMARY: ${summary}
            
            Respond with a JSON object: { "score": number, "rationale": "one sentence explanation" }
        `;

        try {
            const selection = await this.selectRiskModel('general');
            const response = await this.llmService.generateText(
                selection.provider,
                selection.modelId,
                'You are a tormentnexus risk evaluator. Return concise JSON only.',
                prompt,
                {
                    taskType: 'supervisor',
                    routingTaskType: 'general',
                },
            );
            
            // Simple JSON extractor (handles potential markdown wrapping)
            const jsonMatch = response.content.match(/\{.*\}/s);
            const data = jsonMatch ? JSON.parse(jsonMatch[0]) : { score: 50, rationale: "Failed to parse risk analysis." };
            
            const score = Math.min(Math.max(data.score, 0), 100);
            return {
                score,
                level: this.getRiskLevel(score),
                rationale: data.rationale
            };
        } catch (error) {
            console.error("[Agents:Risk] Error calculating risk score:", error);
            return { score: 50, level: 'medium', rationale: "Risk evaluation failed due to service error." };
        }
    }

    /**
     * Evaluates the risk of an implementation plan.
     */
    async evaluatePlanRisk(planText: string): Promise<RiskResult> {
        const prompt = `
            Analyze this implementation plan and provide a risk score (0-100).
            
            PLAN:
            ${planText}
            
            Consider:
            1. Scope: Tightly bounded or sweeping?
            2. Impact: Does it touch auth, security, state, or DB schemas?
            3. Destructiveness: Is it mostly adding or deleting/modifying?

            Respond with a JSON object: { "score": number, "rationale": "one sentence explanation" }
        `;

        try {
            const selection = await this.selectRiskModel('planning');
            const response = await this.llmService.generateText(
                selection.provider,
                selection.modelId,
                'You are a tormentnexus implementation risk evaluator. Return concise JSON only.',
                prompt,
                {
                    taskType: 'supervisor',
                    routingTaskType: 'planning',
                },
            );
            
            const jsonMatch = response.content.match(/\{.*\}/s);
            const data = jsonMatch ? JSON.parse(jsonMatch[0]) : { score: 50, rationale: "Failed to parse plan risk analysis." };
            
            const score = Math.min(Math.max(data.score, 0), 100);
            return {
                score,
                level: this.getRiskLevel(score),
                rationale: data.rationale
            };
        } catch (error) {
            console.error("[Agents:Risk] Error evaluating plan risk:", error);
            return { score: 50, level: 'medium', rationale: "Plan evaluation failed." };
        }
    }

    private getRiskLevel(score: number): RiskResult['level'] {
        if (score <= 20) return 'low';
        if (score <= 50) return 'medium';
        if (score <= 80) return 'high';
        return 'critical';
    }
}
