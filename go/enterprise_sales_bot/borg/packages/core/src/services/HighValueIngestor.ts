/**
 * @file HighValueIngestor.ts
 * @module packages/core/src/services/HighValueIngestor
 *
 * WHAT: Specialized ingestion and processing for high-value external resources.
 * Performs deep semantic extraction and tool discovery for curated links.
 *
 * WHY: High-value links (e.g. popular MCP servers, key documentation) require 
 * more than just a basic crawl. They need to be fully assimilated into 
 * TormentNexus's available toolset and knowledge graph.
 *
 * HOW:
 * 1. Filter links in the backlog by priority/value (stars, domain, manual tag).
 * 2. Run an intensive "Deep Dive" LLM turn on the content.
 * 3. Extract structured features: MCP recipes, prompt templates, skill instructions.
 * 4. Register extracted tools/skills into the live control plane.
 */

import { linksBacklogRepository } from "../db/repositories/links-backlog.repo.js";
import { LLMService } from "@tormentnexus/ai";
import { McpConfigService } from "./McpConfigService.js";
import { SkillRegistry } from "../skills/SkillRegistry.js";

export class HighValueIngestor {
    constructor(
        private llmService: LLMService,
        private mcpConfig: McpConfigService,
        private skillRegistry: SkillRegistry
    ) {}

    /**
     * Scans for high-value links and processes them in-depth.
     */
    public async processHighValueQueue(limit: number = 3) {
        const links = await linksBacklogRepository.listLinks({ 
            limit, 
            research_status: 'done' 
        });

        // Heuristic for "High Value": stars > 100 or specific tags
        const highValue = links.filter(l => {
            const payload = l.raw_payload as any;
            return (payload?.stars && payload.stars > 100) || 
                   (l.tags && l.tags.includes('mcp-server')) ||
                   (l.tags && l.tags.includes('high-value'));
        });

        console.log(`[HighValueIngestor] 💎 Processing ${highValue.length} high-value items...`);

        for (const link of highValue) {
            await this.deepDive(link);
        }
    }

    private async deepDive(link: any) {
        console.log(`[HighValueIngestor] 🔍 Deep diving into: ${link.title || link.url}`);

        const prompt = `
            Analyze this resource for the TormentNexus Control Plane:
            Title: ${link.page_title || link.title}
            Description: ${link.page_description || link.description}
            URL: ${link.url}

            Your goal is to extract technical artifacts for assimilation.
            1. If it's an MCP server, provide a JSON recipe for 'mcp.jsonc'.
            2. If it contains reusable instructions, extract them as a 'Skill' markdown.
            3. If it's a documentation site, provide a concise reference summary.

            Return JSON only:
            {
              "isMcpServer": boolean,
              "mcpRecipe": object | null,
              "isSkill": boolean,
              "skillContent": string | null,
              "summary": "string"
            }
        `;

        try {
            const model = await this.llmService.modelSelector.selectModel({ taskComplexity: 'high' });
            const response = await this.llmService.generateText(model.provider, model.modelId, "High-Value Analysis", prompt);
            
            let analysis;
            try {
                analysis = JSON.parse(response.content.match(/\{[\s\S]*\}/)?.[0] || "{}");
            } catch (e) {
                console.warn(`[HighValueIngestor] Failed to parse analysis for ${link.url}`);
                return;
            }

            if (analysis.isMcpServer && analysis.mcpRecipe) {
                console.log(`[HighValueIngestor] 📦 Discovered MCP Server: ${link.url}`);
                const serverName = (link.page_title || link.title || 'discovered-server').toLowerCase().replace(/\s+/g, '-');
                await this.mcpConfig.addServerConfig(serverName, analysis.mcpRecipe);
            }

            if (analysis.isSkill && analysis.skillContent) {
                console.log(`[HighValueIngestor] 🧠 Discovered Skill: ${link.url}`);
                const skillId = (link.page_title || link.title || 'discovered-skill').toLowerCase().replace(/\s+/g, '-');
                await this.skillRegistry.createSkill(skillId, link.page_title || link.title, analysis.summary);
            }

            // Update status
            await linksBacklogRepository.updateLinkStatus(link.uuid, 'done', 100);

        } catch (e: any) {
            console.error(`[HighValueIngestor] Deep dive failed for ${link.url}: ${e.message}`);
        }
    }
}
