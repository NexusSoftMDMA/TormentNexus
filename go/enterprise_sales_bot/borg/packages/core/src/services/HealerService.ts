import { DEFAULT_OPENROUTER_FREE_MODEL, LLMService } from '@tormentnexus/ai';
import type { MCPServer } from '../MCPServer.js';
import path from 'path';
import fs from 'fs';
import { EventEmitter } from 'events';
import { ShellService } from './ShellService.js';

export interface Diagnosis {
    errorType: string;
    description: string;
    file?: string;
    line?: number;
    suggestedFix: string;
    confidence: number;
}

export interface FixPlan {
    id: string;
    diagnosis: Diagnosis;
    filesToModify: { path: string; content: string }[];
    explanation: string;
}

export interface HealRecord {
    timestamp: number;
    error: string;
    fix: FixPlan;
    success: boolean;
    attempts: number;
}

interface LlmTextResponse {
    text?: string;
    content?: string;
}

function extractJsonObject(text: string): string {
    const fencedMatch = text.match(/```(?:json)?\s*([\s\S]*?)\s*```/i);
    if (fencedMatch?.[1]) {
        return fencedMatch[1].trim();
    }

    const firstBrace = text.indexOf('{');
    const lastBrace = text.lastIndexOf('}');
    if (firstBrace >= 0 && lastBrace > firstBrace) {
        return text.slice(firstBrace, lastBrace + 1).trim();
    }

    return text.trim();
}

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

export class HealerService extends EventEmitter {
    private llm: LLMService;
    private server: MCPServer;
    private shell: ShellService;
    private history: HealRecord[] = [];

    public getHistory() {
        return this.history;
    }

    constructor(llm: LLMService, server: MCPServer, shell: ShellService) {
        super();
        this.llm = llm;
        this.server = server;
        this.shell = shell;
    }

    /**
     * The core autonomous loop: diagnose -> fix -> verify -> retry.
     */
    public async healAndVerify(error: Error | string, context?: string, maxAttempts: number = 3): Promise<boolean> {
        let currentError = error;
        let currentContext = context;
        let attempts = 0;

        console.log(`[HealerService] 🚑 Starting autonomous healing loop (Max Attempts: ${maxAttempts})`);

        while (attempts < maxAttempts) {
            attempts++;
            console.log(`[HealerService] 🔄 Attempt ${attempts}/${maxAttempts}`);

            const diagnosis = await this.analyzeError(currentError, currentContext);
            console.log(`[HealerService] 🔍 Diagnosis: ${diagnosis.description} (Confidence: ${diagnosis.confidence})`);

            if (diagnosis.confidence < 0.6) {
                console.warn("[HealerService] ⚠️ Confidence too low to proceed with auto-fix.");
                break;
            }

            if (!diagnosis.file) {
                console.warn("[HealerService] ⚠️ No culprit file identified.");
                break;
            }

            try {
                const plan = await this.generateFix(diagnosis);
                console.log(`[HealerService] 🛠️ Fix Plan Generated: ${plan.explanation}`);

                const applied = await this.applyFix(plan);
                if (!applied) {
                    console.error("[HealerService] ❌ Failed to write fix to disk.");
                    break;
                }

                console.log("[HealerService] 🧪 Verifying fix...");
                const verificationResult = await this.verifyFix(diagnosis.file);

                if (verificationResult.success) {
                    console.log("[HealerService] ✅ Fix verified successfully!");
                    this.recordHistory(currentError, plan, true, attempts);
                    return true;
                } else {
                    console.warn(`[HealerService] ❌ Verification failed: ${verificationResult.error}`);
                    currentError = verificationResult.error || "Verification failed";
                    currentContext = `Attempted fix: ${plan.explanation}. But verification failed with: ${verificationResult.error}`;
                    // Continue loop with new error
                }
            } catch (e) {
                console.error("[HealerService] ❌ Error during healing step:", e);
                break;
            }
        }

        console.error("[HealerService] ❌ Autonomous healing exhausted or failed.");
        return false;
    }

    public async analyzeError(error: Error | string, context?: string): Promise<Diagnosis> {
        const errorStr = typeof error === 'string' ? error : error.message + '\n' + error.stack;

        const prompt = `
        You are The Healer, an expert debugging agent.
        Analyze the following error and context.
        Provide a diagnosis and a suggested fix.
        
        Error:
        ${errorStr}
        
        Context:
        ${context || 'No additional context.'}
        
        Return JSON format:
        {
            "errorType": "SyntaxError|RuntimeError|LogicError|...",
            "description": "Short explanation",
            "file": "path/to/culprit.ts (if known)",
            "line": 123 (if known),
            "suggestedFix": "Code snippet or description of fix",
            "confidence": 0.0 to 1.0
        }
        `;

        const response = await this.llm.generateText("openrouter", DEFAULT_OPENROUTER_FREE_MODEL, "You are a JSON-only debugging tool.", prompt, {});

        try {
            return JSON.parse(extractJsonObject(extractLlmText(response)));
        } catch (e) {
            console.error("Failed to parse Healer diagnosis", response);
            return {
                errorType: "Unknown",
                description: "Failed to parse LLM diagnosis",
                suggestedFix: "Manual review required",
                confidence: 0
            };
        }
    }

    public async generateFix(diagnosis: Diagnosis): Promise<FixPlan> {
        if (!diagnosis.file) {
            throw new Error("Cannot generate fix without file path.");
        }

        const filePath = path.isAbsolute(diagnosis.file) ? diagnosis.file : path.join(process.cwd(), diagnosis.file);

        let content = '';
        try {
            content = await fs.promises.readFile(filePath, 'utf-8');
        } catch (e) {
            throw new Error(`Failed to read file: ${filePath}`);
        }

        const prompt = `
        You are The Healer.
        Generate a fix for the following file based on the diagnosis.
        
        Diagnosis: ${diagnosis.description}
        Suggested Fix: ${diagnosis.suggestedFix}
        
        File Content:
        ${content}
        
        Return JSON format:
        {
            "explanation": "Why this fix works",
            "newContent": "The entire new file content"
        }
        `;

        const response = await this.llm.generateText("openrouter", DEFAULT_OPENROUTER_FREE_MODEL, "You are a code repair agent. Return only JSON with 'explanation' and 'newContent'.", prompt, {});

        try {
            const result = JSON.parse(extractJsonObject(extractLlmText(response)));
            return {
                id: Math.random().toString(36).substring(7),
                diagnosis,
                filesToModify: [{ path: filePath, content: result.newContent }],
                explanation: result.explanation
            };
        } catch (e) {
            console.error("Failed to parse fix plan", response);
            throw new Error("Failed to generate valid fix plan.");
        }
    }

    public async applyFix(plan: FixPlan): Promise<boolean> {
        try {
            for (const file of plan.filesToModify) {
                await fs.promises.writeFile(file.path, file.content, 'utf-8');
            }
            return true;
        } catch (e) {
            console.error("Failed to apply fix", e);
            return false;
        }
    }

    /**
     * Verifies the fix by running relevant tests or type checking.
     */
    private async verifyFix(culpritFile: string): Promise<{ success: boolean; error?: string }> {
        // Simple heuristic: if it's a TS file, run tsc on it or run vitest if it's a test
        // Better: run 'npm test' or similar if configured.
        // For now, let's try to run the specific test if it exists, or just a generic check.

        const testFile = culpritFile.replace(/\.ts$/, '.test.ts');
        const commands = [];

        if (fs.existsSync(testFile)) {
            commands.push(`npx vitest run ${testFile}`);
        } else {
            // Generic type check as fallback
            commands.push(`npx tsc --noEmit ${culpritFile}`);
        }

        try {
            for (const cmd of commands) {
                await this.shell.execute(cmd);
            }
            return { success: true };
        } catch (e: any) {
            return { success: false, error: e.message || String(e) };
        }
    }

    private recordHistory(error: Error | string, fix: FixPlan, success: boolean, attempts: number) {
        const item: HealRecord = {
            timestamp: Date.now(),
            error: typeof error === 'string' ? error : error.message,
            fix,
            success,
            attempts
        };
        this.history.push(item);
        this.emit('heal', item);
    }
}
