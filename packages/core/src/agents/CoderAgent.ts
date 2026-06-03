import { SpecializedAgent } from '../mesh/SpecializedAgent.js';
import { LLMService } from '@tormentnexus/ai';
import fs from 'fs/promises';
import path from 'path';

/**
 * CoderAgent
 * A specialized agent responsible for writing and modifying code.
 * 
 * Capabilities:
 * - Integrates with `LLMService` to generate code based on tasks.
 * - Parses structured JSON output from the LLM to determine filenames and content.
 * - Writes directly to the filesystem (Phase 62 Ignition Level Capability).
 */
export class CoderAgent extends SpecializedAgent {
    private llm: LLMService;

    constructor(llm: LLMService) {
        super('Coder', ['coding', 'refactoring', 'debugging']);
        this.llm = llm;
    }

    /**
     * Handles a coding task.
     * 1. Constructs a prompt for the LLM.
     * 2. Calls LLM via ModelSelector (High Complexity).
     * 3. Parses the response for a file write plan.
     * 4. Executes the file write operation.
     * 5. Returns a structured result with file changes and reasoning.
     */
    public async handleTask(offer: any): Promise<any> {
        console.log(`[CoderAgent] 🧠 Analyzing task: "${offer.task}"`);

        // 1. Think / Generate Code
        const prompt = `You are a Coder Agent in the TormentNexus Collective.
Task: ${offer.task}

Return JSON with:
{
  "filename": "string",
  "content": "string",
  "reasoning": "string"
}
Output ONLY valid JSON.`;

        try {
            const model = await this.llm.modelSelector.selectModel({ taskComplexity: 'high', routingTaskType: 'coding' });
            const completion = await this.llm.generateText(model.provider, model.modelId, 'You are an expert software engineer.', prompt, {
                taskComplexity: 'high',
                routingTaskType: 'coding',
            });
            const response = completion.content;

            let plan;
            try {
                plan = JSON.parse(response);
            } catch (e) {
                // Fallback if LLM output markdown code block
                const match = response.match(/```json\n([\s\S]*?)\n```/);
                if (match) {
                    plan = JSON.parse(match[1]);
                } else {
                    throw new Error("Failed to parse LLM JSON response");
                }
            }

            console.log(`[CoderAgent] 💡 Generated Plan: Write ${plan.filename}`);

            // 2. Execute Action (Write to Disk)
            // Phase 95: Strict Git Worktree sandbox mapping with fallback
            const workspaceRoot = offer.worktreePath || process.cwd();
            const targetPath = path.resolve(workspaceRoot, plan.filename);

            if (!targetPath.startsWith(workspaceRoot)) {
                throw new Error(`Security Violation: Attempted to write outside the assigned Swarm Worktree isolation zone: ${targetPath}`);
            }

            await fs.writeFile(targetPath, plan.content, 'utf-8');
            console.log(`[CoderAgent] 💾 Wrote to ${targetPath}`);

            return {
                status: 'completed',
                filesChanged: [plan.filename],
                reasoning: plan.reasoning,
                // Phase 96: Execution Telemetry
                modelMetadata: {
                    provider: model.provider,
                    modelId: model.modelId
                }
            };

        } catch (error: any) {
            console.error(`[CoderAgent] 💥 Error:`, error);
            throw error;
        }
    }
}
