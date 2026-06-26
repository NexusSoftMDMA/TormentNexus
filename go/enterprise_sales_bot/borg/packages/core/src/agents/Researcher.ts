
import { MCPServer } from '../MCPServer.js';
import fs from 'fs';
import path from 'path';

interface ResearchTarget {
    name: string;
    url: string;
    description: string;
    category: string;
}

export class Researcher {
    private server: MCPServer;
    private queue: ResearchTarget[] = [];
    private processed: Set<string> = new Set();
    private outputPath: string;

    constructor(server: MCPServer) {
        this.server = server;
        this.outputPath = path.join(process.cwd(), 'packages', 'core', 'data', 'raw_skills');
    }

    async loadQueue(csvPath: string) {
        const content = fs.readFileSync(csvPath, 'utf-8');
        const lines = content.split('\n').slice(1); // Skip header

        for (const line of lines) {
            // Simple parsing (assume no internal commas for MVP)
            const [name, url, desc, category] = line.split(',').map(s => s?.trim());
            if (name && url && url.startsWith('http')) {
                this.queue.push({ name, url, description: desc || '', category: category || 'Uncategorized' });
            }
        }
        console.log(`[Researcher] Loaded ${this.queue.length} targets.`);
    }

    async start(limit: number = 5) {
        console.log(`[Researcher] Starting batch of ${limit}...`);

        let count = 0;
        while (count < limit && this.queue.length > 0) {
            const target = this.queue.shift()!;
            if (this.processed.has(target.url)) continue;

            await this.processTarget(target);
            this.processed.add(target.url);
            count++;
        }
        console.log("[Researcher] Batch complete.");
    }

    private extractJsonObject(text: string): Record<string, unknown> | null {
        const trimmed = text.trim();

        try {
            return JSON.parse(trimmed) as Record<string, unknown>;
        } catch {
            // continue
        }

        const fencedMatch = trimmed.match(/```json\s*([\s\S]*?)```/i);
        if (fencedMatch?.[1]) {
            try {
                return JSON.parse(fencedMatch[1].trim()) as Record<string, unknown>;
            } catch {
                // continue
            }
        }

        const objectMatch = trimmed.match(/\{[\s\S]*\}/);
        if (objectMatch?.[0]) {
            try {
                return JSON.parse(objectMatch[0]) as Record<string, unknown>;
            } catch {
                return null;
            }
        }

        return null;
    }

    private async processTarget(target: ResearchTarget) {
        console.log(`[Researcher] Analyzing: ${target.name} (${target.url})`);

        try {
            let html = "";
            try {
                // Let's use the execute tool method to scrape via page reader or fallback
                const result = await this.server.executeTool("read_url", { url: target.url });
                html = typeof result === 'object' && result !== null && 'content' in result && Array.isArray((result as any).content)
                    ? String((result as any).content[0]?.text || '')
                    : String(result);
            } catch (e) {
                // Fallback for simple testing
                console.warn("[Researcher] Failed to use read_url tool, falling back to native fetch", e);
                const response = await fetch(target.url);
                html = await response.text();
            }
            // Simple text extraction
            const text = html.replace(/<[^>]+>/g, ' ').substring(0, 5000);

            // 2. Synthesize using LLM

            const prompt = `
                I am researching a tool called "${target.name}".
                URL: ${target.url}
                Description: ${target.description}
                
                Content:
                ${text}
                
                Output valid JSON identifying 2-3 MCP tools that this service could provide.
                Schema:
                {
                    "name": "${target.name}",
                    "description": "...",
                    "tools": [
                        { "name": "...", "description": "...", "inputSchema": { ... } }
                    ]
                }
            `;

            const model = await this.server.llmService.modelSelector.selectModel({ taskComplexity: 'high', routingTaskType: 'research' });
            const llmResult = await this.server.llmService.generateText(
                model.provider,
                model.modelId,
                'You are a tool research analyst. Output valid JSON only.',
                prompt,
                {
                    taskComplexity: 'high',
                    routingTaskType: 'research',
                },
            );

            const modelText = llmResult?.content?.trim() || '';
            const parsed = this.extractJsonObject(modelText);

            const skillDef = {
                name: String(parsed?.name ?? target.name),
                description: String(parsed?.description ?? target.description),
                category: target.category,
                tools: Array.isArray(parsed?.tools) && parsed.tools.length > 0
                    ? parsed.tools
                    : [
                        {
                            name: `${target.name.toLowerCase().replace(/[^a-z0-9]/g, '_')}_info`,
                            description: `Get info about ${target.name}`,
                            inputSchema: { type: 'object', properties: {} },
                        },
                    ],
                _source: target.url,
                _researchedAt: new Date().toISOString(),
                _modelOutput: modelText,
            };

            // 3. Save
            const safeName = target.name.replace(/[^a-z0-9]/gi, '_').toLowerCase();
            const catDir = path.join(this.outputPath, target.category.replace(/[^a-z0-9]/gi, '_'));
            if (!fs.existsSync(catDir)) fs.mkdirSync(catDir, { recursive: true });

            fs.writeFileSync(
                path.join(catDir, `${safeName}.json`),
                JSON.stringify(skillDef, null, 2)
            );
            console.log(`[Researcher] Saved ${safeName}.json`);

        } catch (e) {
            console.error(`[Researcher] Failed to process ${target.name}:`, e);
        }
    }
}
