import { DeepResearchService } from './DeepResearchService.js';
import type { MCPServer } from '../MCPServer.js';
import { DEFAULT_OPENROUTER_FREE_MODEL, LLMService } from '@tormentnexus/ai';
import { SkillRegistry } from '../skills/SkillRegistry.js'; // Import
import fs from 'fs/promises';
import path from 'path';
import { z } from 'zod';

function getErrorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
}

export class SkillAssimilationService {
    private skillRegistry: SkillRegistry;
    private llm: LLMService;
    private deepResearch: DeepResearchService;
    private skillsDir: string;

    constructor(skillRegistry: SkillRegistry, llm: LLMService, deepResearch: DeepResearchService) {
        this.skillRegistry = skillRegistry;
        this.llm = llm;
        this.deepResearch = deepResearch;
        this.skillsDir = path.join(process.cwd(), 'packages', 'core', 'src', 'skills');
    }

    /**
     * Learn a new skill from documentation or prompt
     */
    public async assimilate(request: { topic: string; docsUrl?: string; autoInstall?: boolean }): Promise<{ success: boolean; toolName?: string; logs: string[] }> {
        const logs: string[] = [];
        const log = (msg: string) => { console.log(`[Assimilation] ${msg}`); logs.push(msg); };

        log(`Starting assimilation for: ${request.topic}`);

        // 1. Research phase
        log("Phase 1: Researching documentation...");
        let researchContext = "";

        if (request.docsUrl) {
            // Direct URL research
            researchContext = `Documentation from ${request.docsUrl}`;
        } else {
            // General research
            // Use 'researchTopic' instead of create/execute plan
            const result = await this.deepResearch.researchTopic(request.topic, 2) as { summary: string; sources?: Array<{ title: string; url: string }> };

            // Format findings based on research result structure
            researchContext = `SUMMARY:\n${result.summary}\n\nSOURCES:\n${result.sources ? result.sources.map((s) => `- ${s.title}: ${s.url}`).join('\n') : 'No sources'}`;
        }

        log("Phase 2: Generating MCP Tool Code...");

        const systemPrompt = `You are a Senior TypeScript Engineer.
Your goal is to write a robust Model Context Protocol (MCP) Tool in TypeScript.
The tool must be a standalone file that exports a tool definition.

Input Context:
${researchContext}

Constraints:
- Use 'zod' for input schema.
- Export a const named 'toolDefinition' matching { name, description, schema, execute }.
- Or simpler: Export a class implementing 'Tool' interface if that's the project style.
- IMPORTANT: Check existing 'packages/core/src/skills' pattern.
- NO placeholders. Real implementation using 'child_process' or 'fetch'.
`;

        const codeResponse = await this.llm.generateText("openrouter", DEFAULT_OPENROUTER_FREE_MODEL, systemPrompt, `Write an MCP tool for: ${request.topic}`);
        let code = codeResponse.content.replace(/```typescript/g, '').replace(/```/g, '').trim();

        // Basic validation/cleanup of code
        if (!code.includes("export")) {
            log("Error: Generated code has no exports.");
            return { success: false, logs };
        }

        // 3. Save to file
        const safeName = request.topic.toLowerCase().replace(/[^a-z0-9]/g, '_');
        const fileName = `${safeName}.ts`;
        const filePath = path.join(this.skillsDir, fileName);

        log(`Phase 3: Writing to ${fileName}...`);
        await fs.mkdir(this.skillsDir, { recursive: true });
        await fs.writeFile(filePath, code);

        // 4. Hot Reload (Register)
        log("Phase 4: Hot Reloading...");
        try {
            // In a real scenario, we might need to compile via 'tsc' or 'swc' first if not using ts-node/tsx runner
            // For now, we assume the server can load TS or we just validate it.

            // To properly load, we might need to tell SkillRegistry to scan.
            // this.server.skillRegistry.scan();
            log("Tool saved. Restart might be required unless Hot Reload is fully enabled.");
        } catch (e: unknown) {
            log(`Hot reload failed: ${getErrorMessage(e)}`);
        }

        return { success: true, toolName: safeName, logs };
    }
}
