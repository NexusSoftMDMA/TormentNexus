
import dotenv from 'dotenv';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const root = path.resolve(__dirname, '../../../../');

// IMMEDIATE ENV LOADING before other imports!
dotenv.config({ path: path.join(root, '.env') });
dotenv.config({ path: path.join(root, 'packages/core/.env') });

import fs from 'fs/promises';
import { SkillAssimilationService } from '../services/SkillAssimilationService.js';
import { SkillRegistry } from '../skills/SkillRegistry.js';
import { LLMService, ModelSelector } from '@tormentnexus/ai';
import { DeepResearchService } from '../services/DeepResearchService.js';
import { MemoryManager } from '../services/MemoryManager.js';
import { SearchService } from '@tormentnexus/search';

async function run() {
    const indexPath = path.join(root, 'TORMENTNEXUS_MASTER_INDEX.jsonc');
    const skillsRoot = path.join(root, '.tormentnexus', 'skills');

    console.log(`[Assimilator] Root: ${root}`);
    console.log(`[Assimilator] ANTHROPIC_API_KEY present: ${!!process.env.ANTHROPIC_API_KEY}`);

    try {
        const content = await fs.readFile(indexPath, 'utf-8');
        // Safe comment removal that ignores URLs and strings
        const cleanJSON = content.replace(/\\"|"(?:\\"|[^"])*"|(\/\/.*|\/\*[\s\S]*?\*\/)/g, (m, g) => g ? "" : m);
        const index = JSON.parse(cleanJSON);

        const selector = new ModelSelector();
        // Mock LLM for offline assimilation test
        const llm = {
            generateText: async () => ({ content: "export const toolDefinition = { name: 'mock_tool', description: 'Mock Tool', schema: {}, execute: async () => 'Executed' };" }),
            generateJSON: async () => ({}),
            // Add other necessary methods if SkillAssimilationService calls them
        } as unknown as LLMService;
        const registry = new SkillRegistry([skillsRoot]);

        const memoryManager = new MemoryManager(root);
        const searchService = new SearchService();
        const mockServer = { executeTool: async () => ({ content: [{ text: "Mock content" }] }) };
        const deepResearch = new DeepResearchService(mockServer, llm, searchService, memoryManager);

        const assimilator = new SkillAssimilationService(registry, llm, deepResearch);

        const itemsToProcess = [
            ...index.categories.mcp_servers,
            ...index.categories.universal_harness,
            ...index.categories.skills
        ].filter(i => i.status === 'researching' || i.status === 'prioritized' || i.status === 'awaiting_ingest');

        console.log(`[Assimilator] Ingesting ${itemsToProcess.length} items from research index...`);

        for (const item of itemsToProcess) {
            try {
                console.log(`[Assimilator] Processing: ${item.name}`);
                const res = await assimilator.assimilate({
                    topic: item.name,
                    docsUrl: item.url
                });
                console.log(res);
            } catch (e: any) {
                console.error(`[Assimilator] Failed to assimilate ${item.name}: ${e.message}`);
                console.error(e);
            }
        }
    } catch (e: any) {
        console.error(`[Assimilator] Error loading index: ${e.message}`);
    }
}

run();
