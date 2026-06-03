import { EventBus, SystemEvent } from '../services/EventBus.js';
import { DEFAULT_OPENROUTER_FREE_MODEL, LLMService } from '@tormentnexus/ai';
import AgentMemoryService from '../services/AgentMemoryService.js';
import { contextHarvester } from '../services/ContextHarvester.js';
import fs from 'fs/promises';
import path from 'path';

/**
 * MemoryHarvestReactor
 * 
 * Automatically "harvests" context from file system changes.
 * When a file is created or modified, it semantically analyzes the new content
 * and updates TormentNexus's long-term memory graph.
 */
export class MemoryHarvestReactor {
    private eventBus: EventBus;
    private llmService: LLMService;
    private memoryService: AgentMemoryService;
    private isRunning = false;

    constructor(eventBus: EventBus, llmService: LLMService, memoryService: AgentMemoryService) {
        this.eventBus = eventBus;
        this.llmService = llmService;
        this.memoryService = memoryService;
    }

    public start() {
        if (this.isRunning) return;
        this.isRunning = true;

        this.eventBus.onEvent('file:change', (event) => this.handleFileEvent(event));
        this.eventBus.onEvent('file:create', (event) => this.handleFileEvent(event));
        console.log("[MemoryHarvestReactor] 🧠 Sensory harvesting active.");
    }

    private getFileEventPayload(event: SystemEvent): { absolutePath: string; path: string } | null {
        const payload = event.payload;
        if (!payload || typeof payload !== 'object') {
            return null;
        }

        const absolutePath = Reflect.get(payload, 'absolutePath');
        const relativePath = Reflect.get(payload, 'path');

        if (typeof absolutePath !== 'string' || typeof relativePath !== 'string') {
            return null;
        }

        return { absolutePath, path: relativePath };
    }

    private async handleFileEvent(event: SystemEvent) {
        const payload = this.getFileEventPayload(event);
        if (!payload) return;

        const filePath = payload.absolutePath;
        const relativePath = payload.path;

        // Skip non-source files or huge files
        if (!relativePath.match(/\.(ts|tsx|js|jsx|md|py|go|rs|json|yml|yaml)$/i)) return;

        try {
            const content = await fs.readFile(filePath, 'utf-8');
            if (content.length > 100000) return; // Skip massive files for harvesting

            // Harvest the active file immediately into the ephemeral context window!
            // This ensures LLMs instantly know about changes to important files.
            contextHarvester.harvest('active-file', content, {
                path: relativePath,
                timestamp: Date.now()
            });

            const prompt = `
            You are a TormentNexus Knowledge Harvester.
            A file in the repository has been updated: ${relativePath}
            
            Analyze the content and extract the most important architectural rules, 
            exported symbols, or persistent facts that an agent should remember about this file.
            
            Return JSON ONLY:
            {
                "important": boolean,
                "fact": "A concise summary of the file's purpose or a specific rule found within.",
                "tags": ["component", "utility", "rule", etc.]
            }
            `;

            // Use the 'cheapest' strategy for background harvesting
            const response = await this.llmService.generateText(
                'openrouter', 
                DEFAULT_OPENROUTER_FREE_MODEL, 
                'You extract technical knowledge.', 
                prompt,
                { routingStrategy: 'cheapest' }
            );

            // Basic JSON extraction
            const textContent = response.content;
            const start = textContent.indexOf('{');
            const end = textContent.lastIndexOf('}');
            
            if (start !== -1 && end !== -1) {
                const result = JSON.parse(textContent.slice(start, end + 1));
                if (result.important && result.fact) {
                    await this.memoryService.add(result.fact, 'working', 'project', {
                        source: 'file_sensor_harvest',
                        file: relativePath,
                        tags: result.tags || []
                    });
                    
                    this.eventBus.emitEvent('memory:harvested', 'MemoryHarvestReactor', {
                        file: relativePath,
                        fact: result.fact
                    });
                }
            }
        } catch (e) {
            // Silent fail for background harvesting
        }
    }
}
