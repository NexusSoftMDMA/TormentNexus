
import fs from 'fs';
import path from 'path';
import { MCPServer } from '../MCPServer.js';
import { ResearchService } from '../services/ResearchService.js';
import { MemoryManager } from '../services/MemoryManager.js';

/**
 * Batch Process Links from LINKS_TO_PROCESS.md
 */
export async function runResearchPipeline(server: MCPServer) {
    const filePath = path.join(process.cwd(), 'LINKS_TO_PROCESS.md');
    if (!fs.existsSync(filePath)) {
        console.error("LINKS_TO_PROCESS.md not found.");
        return;
    }

    const content = fs.readFileSync(filePath, 'utf-8');
    const lines = content.split('\n');
    const linkRegex = /https?:\/\/[^\s)]+/g;

    const memory = new MemoryManager(process.cwd());
    const research = new ResearchService(server, memory);

    console.log(`[Pipeline] Starting research on ${lines.length} lines...`);

    for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        if (line.includes('[x]')) continue; // Skip processed

        const matches = line.match(linkRegex);
        if (matches) {
            for (const url of matches) {
                console.log(`[Pipeline] Processing: ${url}`);
                try {
                    const result = await research.ingest(url);
                    console.log(`[Pipeline] Result: ${result}`);

                    // Mark as done in file
                    lines[i] = line.replace('[ ]', '[x]');
                    fs.writeFileSync(filePath, lines.join('\n'));
                } catch (e) {
                    console.error(`[Pipeline] Failed ${url}:`, e);
                }
            }
        }
    }
}
