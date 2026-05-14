import { EventBus, SystemEvent } from './EventBus.js';
import { LLMService, DEFAULT_OPENROUTER_FREE_MODEL } from '@borg/ai';
import AgentMemoryService from './AgentMemoryService.js';

export interface ChatInteraction {
    user: string;
    assistant: string;
    timestamp: number;
}

/**
 * PassiveHarvestingService
 *
 * Implements the "Passive Memory Harvesting" pattern by observing agent-user
 * and agent-agent traffic and extracting technical facts into the L2 Vault.
 */
export class PassiveHarvestingService {
    private interactionBuffer: ChatInteraction[] = [];
    private readonly MAX_BUFFER_SIZE = 5;

    constructor(
        private eventBus: EventBus,
        private llmService: LLMService,
        private memoryService: AgentMemoryService
    ) {}

    public start() {
        this.eventBus.subscribe('agent:chat', (event) => this.handleChat(event));
        console.log('[PassiveHarvest] 🛰️ Passive memory harvesting active.');
    }

    private async handleChat(event: SystemEvent) {
        const payload = event.payload as any;
        if (payload?.user && payload?.assistant) {
            this.interactionBuffer.push({
                user: payload.user,
                assistant: payload.assistant,
                timestamp: Date.now()
            });

            if (this.interactionBuffer.length >= this.MAX_BUFFER_SIZE) {
                await this.flush();
            }
        }
    }

    private async flush() {
        const interactions = [...this.interactionBuffer];
        this.interactionBuffer = [];

        const context = interactions.map(i => `User: ${i.user}\nAssistant: ${i.assistant}`).join('\n\n---\n\n');

        const prompt = `
        You are a Borg Passive Harvester.
        Analyze the following chat interactions and extract any durable technical facts,
        architectural decisions, or user preferences that should be remembered.

        Focus on:
        1. "X is preferred over Y"
        2. "The system uses Z for task A"
        3. "New rule: always do B"

        Return JSON list of facts:
        {
            "facts": [
                { "content": "fact string", "tags": ["tag1", "tag2"] }
            ]
        }
        `;

        try {
            const response = await this.llmService.generateText(
                'openrouter',
                DEFAULT_OPENROUTER_FREE_MODEL,
                'You extract durable facts from logs.',
                `Logs:\n${context}\n\n${prompt}`,
                { routingStrategy: 'cheapest' }
            );

            const text = response.content;
            const start = text.indexOf('{');
            const end = text.lastIndexOf('}');
            if (start !== -1 && end !== -1) {
                const result = JSON.parse(text.slice(start, end + 1));
                for (const fact of result.facts || []) {
                    await this.memoryService.add(fact.content, 'working', 'project', {
                        source: 'passive_chat_harvest',
                        tags: fact.tags || []
                    });
                }
            }
        } catch (e) {
            // Passive fail
        }
    }
}
